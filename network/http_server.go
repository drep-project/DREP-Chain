package network

import (
    "net/http"
    "fmt"
    "BlockChainTest/database"
    "strconv"
    "encoding/json"
    "BlockChainTest/bean"
    "math/big"
    "strings"
    "io/ioutil"
)

type Request struct {
    Method string `json:"method"`
    Params string `json:"params"`
}

type Response struct {
    Code string `json:"method"`
    Body string `json:"params"`
}

func GetBlock(w http.ResponseWriter, r *http.Request) {
    block, err := database.GetBlock(1)
    if err != nil{
        errMsg := "error occurred during database.GetBlock"
        fmt.Println(errMsg, ": ", err)
        resp := &Response{Code:"400", Body:errMsg}
        writeResponse(w, resp)
        return
    }
    bytes, err := json.Marshal(block)
    if err != nil{
        errMsg := "error occurred during json.Marshal(block)"
        fmt.Println(errMsg, ": ", err)
        resp := &Response{Code:"400", Body:errMsg}
        writeResponse(w, resp)
        return
    }
    body := string(bytes)
    resp := &Response{Code:"200", Body:body}
    writeResponse(w, resp)
}


//func GetBlocksFrom(w http.ResponseWriter, r *http.Request){
//
//}

func GetAllBlocks(w http.ResponseWriter, _ *http.Request) {
    blocks, err := database.GetAllBlocks()

    if err != nil {
        errMsg := "error occurred during database.GetAllBlocks"
        fmt.Println(errMsg, ": ", err)
        resp := &Response{Code:"400", Body:errMsg}
        writeResponse(w, resp)
        return
    }
    bytes, err := json.Marshal(blocks)
    if err != nil{
        errMsg := "error occurred during json.Marshal(block)"
        fmt.Println(errMsg, ": ", err)
        resp := &Response{Code:"400", Body:errMsg}
        writeResponse(w, resp)
        return
    }
    body := string(bytes)
    resp := &Response{Code:"200", Body:body}
    writeResponse(w, resp)
}

func GetHighestBlock(w http.ResponseWriter, _ *http.Request) {
    block, err := database.GetHighestBlock()
    if err != nil {
        errMsg := "error occurred during database.GetHighestBlock"
        fmt.Println(errMsg, ": ", err)
        resp := &Response{Code:"400", Body:errMsg}
        writeResponse(w, resp)
        return
    }
    bytes, err := json.Marshal(block)
    if err != nil{
        errMsg := "error occurred during json.Marshal(block)"
        fmt.Println(errMsg, ": ", err)
        resp := &Response{Code:"400", Body:errMsg}
        writeResponse(w, resp)
        return
    }
    body := string(bytes)
    resp := &Response{Code:"200", Body:body}
    writeResponse(w, resp)
}

//func PutBlock(w http.ResponseWriter, r *http.Request) {
//
//}

func GetMaxHeight(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var height int64
    if value, ok := params["address"].(int64); ok {
        height = value
    }

    height, err := database.GetMaxHeight()
    if err != nil {
        errMsg := "error occurred during database.GetMaxHeight()"
        fmt.Println(errMsg, ": ", err)
        resp := &Response{Code:"400", Body:errMsg}
        writeResponse(w, resp)
        return
    }

    body := "height:" + strconv.FormatInt(height, 10)
    resp := &Response{Code:"200", Body:body}
    writeResponse(w, resp)
}

func PutMaxHeight(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var height int64
    if value, ok := params["address"].(int64); ok {
        height = value
    }

    err := database.PutMaxHeight(height)
    if err != nil {
        errMsg := "error occurred during database.PutMaxHeight()"
        fmt.Println(errMsg, ": ", err)
        resp := &Response{Code:"400", Body:errMsg}
        writeResponse(w, resp)
        return
    }

    resp := &Response{Code:"200", Body:"[database PutMaxHeight] succeed!"}
    writeResponse(w, resp)
}

func GetBalance(w http.ResponseWriter, r *http.Request) {
    // find param in http.Request
    params := analysisReqParam(r)
    var address string
    if value, ok := params["address"].(string); ok {
        address = value
    }

    if len(address) == 0 {
        resp := &Response{Code:"400", Body:"param format incorrect"}
        writeResponse(w, resp)
        return
    }

    fmt.Println("BalanceAddress: ", address)
    ca := bean.Hex2Address(address)
    //database.PutBalance(ca, big.NewInt(1314))
    //fmt.Println("[database PutBalance] succeed!")

    b, _ := database.GetBalance(ca)
    defer func() {
        if x := recover(); x != nil {
            fmt.Printf("[database GetBalance] caught panic: %v", x)
            resp := &Response{Code:"400", Body:"[database GetBalance] caught panic!"}
            writeResponse(w, resp)
        }
    }()
    body := "balance:" + strconv.FormatInt(b.Int64(), 10)
    resp := &Response{Code:"200", Body:body}
    writeResponse(w, resp)
}

func PutBalance(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)

    var address string
    if value, ok := params["address"].(string); ok {
        address = value
    }

    ca := bean.Hex2Address(address)
    database.PutBalance(ca, big.NewInt(13131313))
    resp := &Response{Code:"400", Body:"database PutBalance] succeed!"}
    writeResponse(w, resp)
}

func GetNonce(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var address string
    if value, ok := params["address"].(string); ok {
        address = value
    }

    if len(address) == 0 {
        resp := &Response{Code:"400", Body:"param format incorrect"}
        writeResponse(w, resp)
        return
    }
    fmt.Println("NonceAddress: ", address)

    ca := bean.Hex2Address(address)
    database.PutNonce(ca, 13131313)

    nonce, _ := database.GetNonce(ca)
    body := "nonce:" + strconv.FormatInt(nonce, 10)
    resp := &Response{Code:"200", Body:body}
    writeResponse(w, resp)
}

func PutNonce(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var address string
    if value, ok := params["address"].(string); ok {
        address = value
    }
    ca := bean.Hex2Address(address)
    database.PutNonce(ca, 13131313)

    resp := &Response{Code:"200", Body:"[database PutNonce] succeed!"}
    writeResponse(w, resp)
}

func GetStateRoot(w http.ResponseWriter, _ *http.Request) {
    b := database.GetStateRoot()
    body := "StateRoot:" + string(b)
    resp := &Response{Code:"200", Body:body}
    writeResponse(w, resp)
}

var methodsMap = map[string] http.HandlerFunc {
    "/GetAllBlocks": GetAllBlocks,
    "/GetBlock": GetBlock,
    "/GetHighestBlock": GetHighestBlock,
    "/GetMaxHeight": GetMaxHeight,
    "/PutMaxHeight": PutMaxHeight,
    "/GetBalance": GetBalance,
    "/PutBalance": PutBalance,
    "/GetNonce": GetNonce,
    "/PutNonce": PutNonce,
    "/GetStateRoot": GetStateRoot,
}

func HttpStart() {
    for pattern, handleFunc := range(methodsMap) {
        http.HandleFunc(pattern, handleFunc)
    }
    fmt.Println("http server is ready for listen port: 8880")
    err := http.ListenAndServe("localhost:8880", nil)
    if err != nil {
        fmt.Println("http listen failed")
    }
}

//func logPanics(handle http.HandlerFunc) http.HandlerFunc {
//    return func(writer http.ResponseWriter, request *http.Request) {
//        defer func() {
//            if x := recover(); x != nil {
//                fmt.Printf("[%v] caught panic: %v", request.RemoteAddr, x)
//            }
//        }()
//        handle(writer, request)
//    }
//}

func writeResponse(w http.ResponseWriter, resp *Response) {
    b, err := json.Marshal(resp)
    if err != nil {
        fmt.Println("error occured resp marshal:", err)
    }
    w.Write(b)
}

func analysisReqParam(r *http.Request) map[string] interface{} {
    params := make(map[string] interface{}, 50)
    switch r.Method {
    case "GET":
        fmt.Println("method: GET")
        r.ParseForm()
        fmt.Println("methodName: ", analysisReqMethodName(r.RequestURI))
        for k, v := range(r.Form) {
            // url.values is a slice of string
            params[k] = v[0]
        }
    case "POST":
        fmt.Println("method: POST")
        result, _ := ioutil.ReadAll(r.Body)
        r.Body.Close()
        fmt.Printf("%s\n", result)
        json.Unmarshal(result, &params)
        //m := f.(map[string]interface{})
        analysisParamsType(params)
    }
    fmt.Println("params: ", params)
    return params
}

func analysisReqMethodName(uri string) (methodName string) {
    s := strings.Split(uri, "?")
    methodName = strings.Trim(s[0], "/")
    return methodName
}

func analysisParamsType(params map[string] interface{})  {
    for k, v := range params {
        switch vType := v.(type) {
        case string:
            fmt.Println(k, "is string", vType)
        case int:
            fmt.Println(k, "is int", vType)
        case float64:
            fmt.Println(k, "is float64", vType)
        case []interface{}:
            fmt.Println(k, "is an array:")
            for i, u := range vType {
                fmt.Println(i, u)
            }
        default:
            fmt.Println(k, "is an unkown Type to handle")
        }
    }
}

//func analysisGetReqParamWithUri(uri string) (methodName string, params map[string] string)  {
//    s := strings.Split(uri, "?")
//    p := s[1]
//    parts := strings.Split(p, "&")
//
//    dict := make(map[string] string, 50)
//    for _, param := range parts {
//        argument := strings.Split(param, "=")
//
//        key := argument[0]
//        value := argument[1]
//        dict[key] = value
//    }
//    name := strings.Trim(s[0], "/")
//    fmt.Println("MethodName: ", name)
//
//    jsonBytes, _ := json.Marshal(dict)
//    j := string(jsonBytes)
//    fmt.Printf("Params:\n %s\n", j)
//    return name, dict
//}
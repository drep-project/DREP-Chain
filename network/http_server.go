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

//
//func GetBlocksFrom(w http.ResponseWriter, r *http.Request){
//
//}
//
func GetAllBlocks(w http.ResponseWriter, r *http.Request) {
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

func GetHighestBlock(w http.ResponseWriter, r *http.Request) {
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
    height, _ := database.GetMaxHeight()
    body := "height:" + strconv.FormatInt(height, 10)
    resp := &Response{Code:"200", Body:body}
    writeResponse(w, resp)
}

//func PutMaxHeight(w http.ResponseWriter, r *http.Request) {
//
//}

func GetBalance(w http.ResponseWriter, r *http.Request) {
    // find param string in uri
    //_, params := analysisGetReqParamWithUri(r.RequestURI)
    var address string
    r.ParseForm()
    if r.Method == "GET" {
        fmt.Println("method:", r.Method)
        fmt.Println("params", r.Form["address"])
        address = r.Form["address"][0]
    } else if r.Method == "POST" {
        result, _ := ioutil.ReadAll(r.Body)
        r.Body.Close()
        fmt.Printf("%s\n", result)

        var params map[string] interface{}
        json.Unmarshal(result, &params)
        //m := f.(map[string]interface{})
        analysisParamsType(params)
        if value, ok := params["address"].(string); ok {
            address = value
        }
    }

    if len(address) == 0 {
        resp := &Response{Code:"400", Body:"param format incorrect"}
        writeResponse(w, resp)
        return
    }

    fmt.Println("BalanceAddress: ", address)
    ca := bean.Hex2Address(address)
    database.PutBalance(ca, big.NewInt(1314))

    b, _ := database.GetBalance(ca)
    body := "balance:" + strconv.FormatInt(b.Int64(), 10)
    resp := &Response{Code:"200", Body:body}
    writeResponse(w, resp)
}

func PutBalance(w http.ResponseWriter, r *http.Request) {
   address := "1A2B3C"
   ca := bean.Hex2Address(address)
   database.PutBalance(ca, big.NewInt(13131313))
}

func GetNonce(w http.ResponseWriter, r *http.Request) {
    _, params := analysisGetReqParamWithUri(r.RequestURI)
    address := params["address"]
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

//func PutNonce(w http.ResponseWriter, r *http.Request) {
//       address := "1A2B3C"
//       ca := bean.Hex2Address(address)
//       database.PutNonce(ca, 13131313)
//}

func GetStateRoot(w http.ResponseWriter, r *http.Request) {
    b := database.GetStateRoot()
    body := "StateRoot:" + string(b)
    resp := &Response{Code:"200", Body:body}
    writeResponse(w, resp)
}

func HttpStart() {
    http.HandleFunc("/GetAllBlocks", GetAllBlocks)
    http.HandleFunc("/GetBlock", GetBlock)
    http.HandleFunc("/GetMaxHeight", GetMaxHeight)
    http.HandleFunc("/GetBalance", GetBalance)
    http.HandleFunc("/GetNonce", GetNonce)
    http.HandleFunc("/GetStateRoot", GetStateRoot)

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

func analysisGetReqParamWithUri(uri string) (methodName string, params map[string] string)  {
    s := strings.Split(uri, "?")
    p := s[1]
    parts := strings.Split(p, "&")

    dict := make(map[string] string, 50)
    for _, param := range parts {
        argument := strings.Split(param, "=")

        key := argument[0]
        value := argument[1]
        dict[key] = value
    }
    name := strings.Trim(s[0], "/")
    fmt.Println("MethodName: ", name)

    jsonBytes, _ := json.Marshal(dict)
    json := string(jsonBytes)
    fmt.Printf("Params:\n %s\n", json)
    return name, dict
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
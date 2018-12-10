package rest

import (
    "net/http"
    "fmt"
    "net"
    "BlockChainTest/database"
    "strconv"
    "encoding/json"
    "math/big"
    "strings"
    "io/ioutil"
    "BlockChainTest/log"
    "BlockChainTest/node"
    "BlockChainTest/accounts"
    "github.com/spf13/viper"
)

type Request struct {
    Method string `json:"method"`
    Params string `json:"params"`
}

type Response struct {
    Success bool `json:"success"`
    ErrorMsg string `json:"errMsg"`
    Data interface{} `json:"data"`
}

func GetAllBlocks(w http.ResponseWriter, _ *http.Request) {
    fmt.Println("get all blocks running")
    blocks := database.GetAllBlocks()

    if blocks == nil || len(blocks) == 0 {
        errMsg := "error occurred during database.GetAllBlocks"
        fmt.Println(errMsg)
        resp := &Response{Success:false, Data:errMsg}
        writeResponse(w, resp)
        return
    }

    var body []*BlockWeb
    for _, block := range(blocks) {
        item := ParseBlock(block)
        body = append(body, item)
    }
    resp := &Response{Success:true, Data:body}
    writeResponse(w, resp)
}

func GetBlock(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var height int64
    if value, ok := params["height"].(string); ok {
        height, _ = strconv.ParseInt(value, 10, 64)
    }

    block := database.GetBlock(height)
    if block == nil {
        errMsg := "block is nil"
        fmt.Println(errMsg)
        resp := &Response{Success:false, Data:errMsg}
        writeResponse(w, resp)
        return
    }
    blockWeb := ParseBlock(block)
    resp := &Response{Success:true, Data:blockWeb}
    writeResponse(w, resp)
}

func GetHighestBlock(w http.ResponseWriter, _ *http.Request) {
    block := database.GetHighestBlock()
    if block == nil {
        errMsg := "error occurred during database.GetHighestBlock"
        fmt.Println(errMsg)
        resp := &Response{Success:false, Data:errMsg}
        writeResponse(w, resp)
        return
    }
    blockWeb := ParseBlock(block)
    resp := &Response{Success:true, Data:blockWeb}
    writeResponse(w, resp)
}

func GetMaxHeight(w http.ResponseWriter, _ *http.Request) {
    height := database.GetMaxHeight()
    if height == -1 {
        errMsg := "error occurred during database.GetMaxHeight()"
        fmt.Println(errMsg)
        resp := &Response{Success:false, Data:errMsg}
        writeResponse(w, resp)
        return
    }
    resp := &Response{Success:true, Data:height}
    writeResponse(w, resp)
}

func GetBlocksFrom(w http.ResponseWriter, r *http.Request){
    params := analysisReqParam(r)
    var start, size int64
    if value, ok := params["start"].(string); ok {
        start, _ = strconv.ParseInt(value, 10, 64)
    }
    if value, ok := params["size"].(string); ok {
        size, _ = strconv.ParseInt(value, 10, 64)
    }

    blocks := database.GetBlocksFrom(start, size)
    var body []*BlockWeb
    for _, block := range(blocks) {
        item := ParseBlock(block)
        body = append(body, item)
    }
    resp := &Response{Success:true, Data:body}
    writeResponse(w, resp)
}

func GetBalance(w http.ResponseWriter, r *http.Request) {
    // find param in http.Request
    params := analysisReqParam(r)
    var address string
    var chainId int64
    if value, ok := params["address"].(string); ok {
        address = value[2:]
    }
    if value, ok := params["chainId"].(string); ok {
        chainId, _ = strconv.ParseInt(value, 10, 64)
    }

    if len(address) == 0 {
        resp := &Response{Success:false, ErrorMsg:"param format incorrect"}
        writeResponse(w, resp)
        return
    }

    fmt.Println("BalanceAddress: ", address)
    ca := accounts.Hex2Address(address)
    //database.PutBalance(ca, big.NewInt(1314))
    //fmt.Println("[database PutBalance] succeed!")

    b := database.GetBalanceOutsideTransaction(ca, chainId)
    resp := &Response{Success:true, Data:b.String()}
    writeResponse(w, resp)
}

func GetNonce(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var address string
    var chainId int64
    if value, ok := params["address"].(string); ok {
        address = value
    }
    if value, ok := params["chainId"].(string); ok {
        chainId, _ = strconv.ParseInt(value, 10, 64)
    }

    if len(address) == 0 {
        resp := &Response{Success:false, ErrorMsg:"param format incorrect"}
        writeResponse(w, resp)
        return
    }
    fmt.Println("NonceAddress: ", address)

    ca := accounts.Hex2Address(address)

    nonce := database.GetNonceOutsideTransaction(ca, chainId)
    resp := &Response{Success:true, Data:nonce}
    writeResponse(w, resp)
}

func SendTransaction(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var to string
    var amount string
    var destChain int64
    if value, ok := params["to"].(string); ok {
        to = value
    }
    if value, ok := params["amount"].(string); ok {
        amount = value
    }
    if value, ok := params["destChain"].(string); ok {
        destChain, _ = strconv.ParseInt(value, 10, 64)
    }

    a, succeed := new(big.Int).SetString(amount, 10)
    if succeed == false {
        errorMsg := "params amount parsing error"
        resp := &Response{Success:true, ErrorMsg:errorMsg}
        writeResponse(w, resp)
        return
    }

    t := node.GenerateBalanceTransaction(to, destChain, a)

    var body string
    if node.SendTransaction(t) != nil {
        body = "Offline"
    } else {
        body = "Send finish"
    }

    resp := &Response{Success:true, Data:body}
    writeResponse(w, resp)
}

func GetTransactionsFromBlock(w http.ResponseWriter, r *http.Request) {
   params := analysisReqParam(r)
   var height int64
   if value, ok := params["height"].(string); ok {
       height, _ = strconv.ParseInt(value, 10, 64)
   }
   block := database.GetBlock(height)
   txs := block.Data.TxList
   var txsWeb []*TransactionWeb
   for _, tx := range(txs) {
       tx := ParseTransaction(tx)
       txsWeb = append(txsWeb, tx)
   }
   resp := &Response{Success:true, Data:txsWeb}
   writeResponse(w, resp)
}

func GetReputation(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var address string
    var chainId int64
    if value, ok := params["address"].(string); ok {
        address = value[2:]
    }
    if value, ok := params["chainId"].(string); ok {
        chainId, _ = strconv.ParseInt(value, 10, 64)
    }

    if len(address) == 0 {
        resp := &Response{Success:false, ErrorMsg:"param format incorrect"}
        writeResponse(w, resp)
        return
    }

    ca := accounts.Hex2Address(address)
    b:= database.GetReputationOutsideTransaction(ca, chainId)
    resp := &Response{Success:true, Data:b.String()}
    if (b.Int64() == 0) {
        defaultRep := viper.GetInt64("default_rep")
        fmt.Println("default reputation is :", defaultRep)
        database.PutReputationOutSideTransaction(ca, chainId, big.NewInt(defaultRep))
        resp.Data = viper.GetString("default_rep")
    }
    writeResponse(w, resp)
}

var methodsMap = map[string] http.HandlerFunc {
    "/GetAllBlocks": GetAllBlocks,
    "/GetBlock": GetBlock,
    "/GetHighestBlock": GetHighestBlock,
    "/GetMaxHeight": GetMaxHeight,
    "/GetBlocksFrom": GetBlocksFrom,
    "/GetBalance": GetBalance,
    "/GetNonce": GetNonce,
    "/SendTransaction": SendTransaction,
    "/GetReputation": GetReputation,
    "/GetTransactionsFromBlock": GetTransactionsFromBlock,
}

func HttpStart(restEndPoint string) (net.Listener, error){
    for pattern, handleFunc := range (methodsMap) {
        http.HandleFunc(pattern, handleFunc)
    }

    listen, err := net.Listen("tcp", restEndPoint)
    if err != nil {
        log.Error("start reset server errpr :", err.Error())
        return nil, err
    }
    log.Info("rest server is ready for listen，","endpoint" ,restEndPoint)

    svr := http.Server{}
    go svr.Serve(listen)
    return listen, nil
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
        //fmt.Println("method: GET")
        r.ParseForm()
        //fmt.Println("methodName: ", analysisReqMethodName(r.RequestURI))
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
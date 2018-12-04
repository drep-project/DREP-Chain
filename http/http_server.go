package http

import (
    "net/http"
    "fmt"
    "BlockChainTest/database"
    "strconv"
    "encoding/json"
    "strings"
    "io/ioutil"
)

type Request struct {
    Method string `json:"method"`
    Params string `json:"params"`
}

type Response struct {
    Success bool `json:"success"`
    ErrorMsg string `json:"errMsg"`
    Body interface{} `json:"body"`
}

func GetAllBlocks(w http.ResponseWriter, _ *http.Request) {
    //TODO:加参数，矿工地址
    fmt.Println("get all blocks running")
    blocks := database.GetAllBlocks()

    if blocks == nil || len(blocks) == 0 {
        errMsg := "error occurred during database.GetAllBlocks"
        fmt.Println(errMsg)
        resp := &Response{Success:false, Body:errMsg}
        writeResponse(w, resp)
        return
    }

    bytes, err := json.Marshal(blocks)
    if err != nil{
        errMsg := "error occurred during json.Marshal(block)"
        fmt.Println(errMsg, ": ", err)
        resp := &Response{Success:false, Body:errMsg}
        writeResponse(w, resp)
        return
    }
    body := string(bytes)
    resp := &Response{Success:true, Body:body}
    writeResponse(w, resp)
}

func GetBlock(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var height int64
    if value, ok := params["height"].(int64); ok {
        height = value
    }

    block := database.GetBlock(height)
    if block == nil {
        errMsg := "error occurred during database.GetBlock"
        fmt.Println(errMsg)
        resp := &Response{Success:false, Body:errMsg}
        writeResponse(w, resp)
        return
    }
    bytes, err := json.Marshal(block)
    if err != nil{
        errMsg := "error occurred during json.Marshal(block)"
        fmt.Println(errMsg, ": ", err)
        resp := &Response{Success:false, Body:errMsg}
        writeResponse(w, resp)
        return
    }
    body := string(bytes)
    resp := &Response{Success:true, Body:body}
    writeResponse(w, resp)
}

func GetHighestBlock(w http.ResponseWriter, _ *http.Request) {
    block := database.GetHighestBlock()
    if block == nil {
        errMsg := "error occurred during database.GetHighestBlock"
        fmt.Println(errMsg)
        resp := &Response{Success:false, Body:errMsg}
        writeResponse(w, resp)
        return
    }
    bytes, err := json.Marshal(block)
    if err != nil{
        errMsg := "error occurred during json.Marshal(block)"
        fmt.Println(errMsg, ": ", err)
        resp := &Response{Success:false, Body:errMsg}
        writeResponse(w, resp)
        return
    }
    body := string(bytes)
    resp := &Response{Success:true, Body:body}
    writeResponse(w, resp)
}

func GetMaxHeight(w http.ResponseWriter, _ *http.Request) {
    height := getMaxHeight()
    if height == -1 {
        errMsg := "error occurred during database.GetMaxHeight()"
        resp := &Response{Success:false, Body:errMsg}
        writeResponse(w, resp)
        return
    }
    body := strconv.FormatInt(height, 10)
    resp := &Response{Success:true, Body:body}
    writeResponse(w, resp)
}

func GetBlocksFrom(w http.ResponseWriter, r *http.Request){
    params := analysisReqParam(r)
    var start, size int64
    if value, ok := params["start"].(int64); ok {
        start = value
    }
    if value, ok := params["size"].(int64); ok {
        size = value
    }

    blocks := database.GetBlocksFrom(start, size)
    if blocks == nil || len(blocks) == 0 {
        errMsg := "error occurred during GetBlocksFrom"
        fmt.Println(errMsg)
        resp := &Response{Success:false, Body:errMsg}
        writeResponse(w, resp)
        return
    }
    bytes, err := json.Marshal(blocks)
    if err != nil{
        errMsg := "error occurred during json.Marshal(block)"
        fmt.Println(errMsg, ": ", err)
        resp := &Response{Success:false, Body:errMsg}
        writeResponse(w, resp)
        return
    }
    body := string(bytes)
    resp := &Response{Success:true, Body:body}
    writeResponse(w, resp)
}

func GetBalance(w http.ResponseWriter, r *http.Request) {
    // find param in http.Request
    params := analysisReqParam(r)
    var address string
    var chainId int64
    if value, ok := params["address"].(string); ok {
        address = value
    }
    if value, ok := params["chainId"].(int64); ok {
        chainId = value
    }

    if len(address) == 0 {
        resp := &Response{Success:false, ErrorMsg:"param format incorrect"}
        writeResponse(w, resp)
        return
    }

    balance, err := getBalance(address, chainId)
    if err != nil {
        resp := &Response{Success:false, ErrorMsg:err.Error()}
        writeResponse(w, resp)
        return
    }
    resp := &Response{Success:true, Body:balance.String()}
    writeResponse(w, resp)
}

func GetNonce(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var address string
    var chainId int64
    if value, ok := params["address"].(string); ok {
        address = value
    }
    if value, ok := params["chainId"].(int64); ok {
        chainId = value
    }

    if len(address) == 0 {
        resp := &Response{Success:false, ErrorMsg:"param format incorrect"}
        writeResponse(w, resp)
        return
    }

    nonce, err := getNonce(address, chainId)
    if err != nil {
        resp := &Response{Success:false, ErrorMsg:err.Error(), Body:false}
        writeResponse(w, resp)
        return
    }
    resp := &Response{Success:true, Body:nonce}
    writeResponse(w, resp)
}

func GetStateRoot(w http.ResponseWriter, _ *http.Request) {
    b := database.GetDB().GetStateRoot()
    body := string(b)
    resp := &Response{Success:true, Body:body}
    writeResponse(w, resp)
}

func SendTransferTransaction(w http.ResponseWriter, r *http.Request) {
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
    if value, ok := params["destChain"].(int64); ok {
        destChain = value
    }
    err := sendTransferTransaction(to, amount, destChain)
    if err != nil {
        resp := &Response{Success:false, ErrorMsg:err.Error(), Body:false}
        writeResponse(w, resp)
        return
    }
    resp := &Response{Success:true}
    writeResponse(w, resp)
}

func SendCreateContractTransaction(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var code string
    if value, ok := params["code"].(string); ok {
        code = value
    }
    err := sendCreateContractTransaction(code)
    if err != nil {
        resp := &Response{Success:false, ErrorMsg:err.Error(), Body:false}
        writeResponse(w, resp)
        return
    }
    resp := &Response{Success:true}
    writeResponse(w, resp)
}

func SendCallContractTransaction(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var addr string
    if value, ok := params["address"].(string); ok {
        addr = value
    }
    var chainId int64
    if value, ok := params["chainId"].(int64); ok {
        chainId = value
    }
    var input string
    if value, ok := params["input"].(string); ok {
        input = value
    }
    var readOnly bool
    if value, ok := params["readOnly"].(bool); ok {
        readOnly = value
    }
    err := sendCallContractTransaction(addr, chainId, input, readOnly)
    if err != nil {
        resp := &Response{Success:false, ErrorMsg:err.Error(), Body:false}
        writeResponse(w, resp)
        return
    }
    resp := &Response{Success:true}
    writeResponse(w, resp)
}

func CreateAccount(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var chainId int64
    var keystore string
    if value, ok := params["chainId"].(int64); ok {
        chainId = value
    }
    if value, ok := params["keystore"].(string); ok {
        keystore = value
    }
    hexStr, err := createAccount(chainId, keystore)
    var resp *Response
    if err != nil {
        resp = &Response{Success:false, ErrorMsg:err.Error()}
        writeResponse(w, resp)
        return
    }
    resp = &Response{Success:true, Body:hexStr}
    writeResponse(w, resp)
}

func GetAccount(w http.ResponseWriter, _ *http.Request) {
    account := getAccount()
    resp := &Response{Success:true, Body:account}
    writeResponse(w, resp)
}

func GetMostRecentBlocks(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var n int64
    if value, ok := params["n"].(int64); ok {
        n = value
    }
    blocks := database.GetMostRecentBlocks(n)
    resp := &Response{Success:true, Body:blocks}
    writeResponse(w, resp)
}

func GetTransactionsFormBlock(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var height int64
    if value, ok := params["height"].(int64); ok {
        height = value
    }
    block := database.GetBlock(height)
    txs := block.Data.TxList
    var body []*TransactionWeb
    for _, tx := range(txs) {
        tx := ParseTransaction(tx)
        body = append(body, tx)
    }
    resp := &Response{Success:true, Body:body}
    writeResponse(w, resp)
}

func SetChain(w http.ResponseWriter, r *http.Request) {
    params := analysisReqParam(r)
    var chainId int64
    var dataDir string
    if value, ok := params["chainId"].(int64); ok {
        chainId = value
    }
    if value, ok := params["dataDir"].(string); ok {
        dataDir = value
    }
    err := setChain(chainId, dataDir)
    if err != nil {
        resp := &Response{Success: false, ErrorMsg: err.Error()}
        writeResponse(w, resp)
    }
    resp := &Response{Success: true}
    writeResponse(w, resp)
}

var methodsMap = map[string] http.HandlerFunc {
    "/GetAllBlocks":             GetAllBlocks,
    "/GetBlock":                 GetBlock,
    "/GetHighestBlock":          GetHighestBlock,
    "/GetMaxHeight":             GetMaxHeight,
    "/GetBlocksFrom":            GetBlocksFrom,
    "/GetBalance":               GetBalance,
    "/GetNonce":                 GetNonce,
    "/GetStateRoot":             GetStateRoot,
    "/SendTransferTransaction":  SendTransferTransaction,
    "/SendCreateContractTransaction": SendCreateContractTransaction,
    "/SendCallContractTransaction": SendCallContractTransaction,
    "/CreateAccount":            CreateAccount,
    "/GetAccount":               GetAccount,
    "/GetMostRecentBlocks":      GetMostRecentBlocks,
    "/GetTransactionsFormBlock": GetTransactionsFormBlock,
    "/SetChain":                 SetChain,
}

func HttpStart() {
    go func() {
        for pattern, handleFunc := range (methodsMap) {
            http.HandleFunc(pattern, handleFunc)
        }
        fmt.Println("http server is ready for listen port: 8880")
        err := http.ListenAndServe("localhost:8880", nil)
        if err != nil {
            fmt.Println("http listen failed")
        }
    }()
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
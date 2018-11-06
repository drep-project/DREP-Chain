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
)

type Request struct {
    Method string `json:"method"`
    Params string `json:"params"`
}

type Response struct {
    Code string `json:"method"`
    Body string `json:"params"`
}

//func GetBlock(w http.ResponseWriter, r *http.Request) {
//
//}
//
//func GetBlocksFrom(w http.ResponseWriter, r *http.Request){
//
//}
//
//func GetAllBlocks(w http.ResponseWriter, r *http.Request) {
//
//}
//
//func GetHighestBlock(w http.ResponseWriter, r *http.Request) {
//
//}

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
    _, params := analysisParamWithUri(r.RequestURI)

    address := params["address"]
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

//func PutBalance(w http.ResponseWriter, r *http.Request) {
//    address := "1A2B3C"
//    ca := bean.Hex2Address(address)
//    database.PutBalance(ca, big.NewInt(13131313))
//}

func GetNonce(w http.ResponseWriter, r *http.Request) {
    address := bean.CommonAddress{}
    nonce, _ := database.GetNonce(address)
    body := "nonce:" + strconv.FormatInt(nonce, 10)
    resp := &Response{Code:"200", Body:body}
    writeResponse(w, resp)
}

//func PutNonce(w http.ResponseWriter, r *http.Request) {
//
//}

func GetStateRoot(w http.ResponseWriter, r *http.Request) {
    b := database.GetStateRoot()
    body := "StateRoot:" + string(b)
    resp := &Response{Code:"200", Body:body}
    writeResponse(w, resp)
}

func HttpStart() {
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

func analysisParamWithUri(uri string) (methodName string, params map[string] string)  {
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
package cmd

import (
    "net/http"
    "io/ioutil"
    "strconv"
    "encoding/json"
)

var server = "http://localhost:8880/"

type Response struct {
    Success bool `json:"success"`
    Body interface{} `json:"body"`
    ErrorMsg string `json:"errMsg"`
}

func GetResponse(url string) (*Response, error) {
    res, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    data, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return nil, err
    }
    resp := &Response{}
    err = json.Unmarshal(data, resp)
    if err != nil {
        return nil, err
    }
    return resp, nil
}

func urlBalance(address, chainId string) string {
    return server + "GetBalance?address=" + address + "&chainId=" + chainId
}

func urlNonce(address, chainId string) string {
    return server + "GetNonce?address=" + address + "&chainId=" + chainId
}

func urlMaxHeight() string {
    return server + "GetMaxHeight"
}

func urlBlock(height int64) string {
    return server + "GetBlock?height=" + strconv.FormatInt(height, 10)
}

func urlBlocksFrom(start, size int64) string {
    return server + "GetBlocksFrom?start=" + strconv.FormatInt(start, 10) + "&size=" + strconv.FormatInt(size, 10)
}

func urlMostRecentBlocks(n int64) string {
    return server + "GetMostRecentBlocks?n=" + strconv.FormatInt(n, 10)
}

func urlAllBlocks() string {
    return server + "GetAllBlocks"
}

func urlCreateAccount(chainId int64, keystore string) string {
    return server + "CreateAccount?chainId=" + strconv.FormatInt(chainId, 10) + "&keystore=" + keystore
}

func urlGetAccount() string {
    return server + "GetAccount"
}

func urlSendTransferTransaction(to, destChain, amount string) string {
    return server + "SendTransferTransaction?to=" + to + "&destChain=" + destChain + "&amount=" + amount
}

func urlSendCreateContractTransaction(codeFile string) string {
    return server + "SendCreateContractTransaction?codeFile=" + codeFile
}

func urlSetChain(chainId, dataDir string) string {
    return server + "SetCurrentChain?chainId=" + chainId + "&dataDir=" + dataDir
}
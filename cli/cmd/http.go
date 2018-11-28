package cmd

import (
    "net/http"
    "io/ioutil"
    "strconv"
)

var server = "http://localhost:8880/"

type Response struct {
    Code string `json:"code"`
    Body interface{} `json:"body"`
    ErrorMsg string `json:"errMsg"`
}

func (resp *Response) OK() bool {
    return resp.Code == "200" && resp.ErrorMsg == ""
}

func GetRequest(url string) ([]byte, error) {
    res, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    data, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return nil, err
    }
    return data, nil
}

func urlAccounts() string {
    return server + "GetAccounts"
}

func urlBalance(addr string) string {
    return server + "GetBalance?address=" + addr
}

func urlNonce(addr string) string {
    return server + "GetNonce?address=" + addr
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

func urlCreateAccount() string {
    return server + "CreateAccount"
}

func urlSwitchAccount(addr string) string {
    return server + "SwitchAccount?address=" + addr
}

func urlCurrentAccount() string {
    return server + "CurrentAccount"
}

func urlSendTransaction(to, amount string) string {
    return server + "SendTransaction?to=" + to + "&amount=" + amount
}
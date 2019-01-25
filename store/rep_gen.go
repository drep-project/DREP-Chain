package store

import (
    "BlockChainTest/accounts"
    "encoding/json"
)

var (
    eric = accounts.Bytes2Address([]byte("eric")).Hex()
    sai  = accounts.Bytes2Address([]byte("sai")).Hex()
    long = accounts.Bytes2Address([]byte("long")).Hex()
    xie  = accounts.Bytes2Address([]byte("xie")).Hex()
)

func generateInc(height int64) []byte {
    increments := make([]map[string] interface{}, 4)

    increments[0] = make(map[string] interface{})
    increments[0]["Addr"] = eric
    increments[0]["Day"] = int(height)
    increments[0]["Gain"] = 10 * (int(height % 5) + 1) + 2

    increments[1] = make(map[string] interface{})
    increments[1]["Addr"] = sai
    increments[1]["Day"] = int(height)
    increments[1]["Gain"] = 10 * (int(height % 5) + 1) + 5

    increments[2] = make(map[string] interface{})
    increments[2]["Addr"] = long
    increments[2]["Day"] = int(height)
    increments[2]["Gain"] = 10 * (int(height % 5) + 1) + 8

    increments[3] = make(map[string] interface{})
    increments[3]["Addr"] = xie
    increments[3]["Day"] = int(height)
    increments[3]["Gain"] = 10 * (int(height % 5) + 1) + 11

    data, _ := json.Marshal(increments)
    return data
}

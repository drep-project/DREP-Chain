package config

import (
    "BlockChainTest/mycrypto"
    "encoding/hex"
)

type PK struct {
    X string
    Y string
}

func FormatPubKey(pubKey *mycrypto.Point) *PK {
    return &PK{
        X: hex.EncodeToString(pubKey.X),
        Y: hex.EncodeToString(pubKey.Y),
    }
}

func ParsePK(pk *PK) *mycrypto.Point {
    x, _ := hex.DecodeString(pk.X)
    y, _ := hex.DecodeString(pk.Y)
    return &mycrypto.Point{
        X: x,
        Y: y,
    }
}

type DebugNode struct {
    PubKey  *PK
    Address string
    IP      string
    Port    int
}

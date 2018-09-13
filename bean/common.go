package bean

import (
    "BlockChainTest/mycrypto"
    "encoding/hex"
)

const (
    MsgTypeBlockHeader = 1
    MsgTypeBlock       = 3
    MsgTypeTransaction = 4
    MsgTypeSetUp       = 5
    MsgTypeCommitment  = 6
    MsgTypeChallenge   = 7
    MsgTypeResponse    = 8
    MsgTypeNewPeer     = 9
    MsgTypePeerList    = 10
    MsgTypeMinerInfo   = 11
)

const (
    MINER    = 0
    OTHER     = 1
)

type Address string

func (addr Address) String() string {
    return string(addr)
}

func Addr(pubKey *mycrypto.Point) Address {
    j := pubKey.Bytes()
    h := mycrypto.Hash256(j)
    str := hex.EncodeToString(h[len(h) - AddressLen:])
    return Address(str)
}
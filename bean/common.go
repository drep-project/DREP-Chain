package bean

import (
    "BlockChainTest/mycrypto"
    "encoding/hex"
)

const (
    MsgTypeAccount     = 0
    MsgTypeBlockHeader = 1
    MsgTypeBlockData   = 2
    MsgTypeBlock       = 3
    MsgTypeTransaction = 4
    MsgTypeSetUp       = 5
    MsgTypeCommitment  = 6
    MsgTypeChallenge   = 7
    MsgTypeResponse    = 8
    MsgTypeNewPeer     = 9
    MsgTypePeerList    = 10
)

const (
    LEADER    = 0
    MEMBER    = 1
    OTHER     = 2
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
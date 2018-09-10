package bean

import (
    "BlockChainTest/crypto"
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
    MsgTypeNewComer    = 9
    MsgTypeUser        = 10
)

const (
    LEADER    = 0
    MEMBER    = 1
    NEWCOMER  = 2
    OTHER     = 3
)

type Address string

func (addr Address) String() string {
    return string(addr)
}

func Addr(pubKey *crypto.Point) Address {
    j := pubKey.Bytes()
    h := crypto.Hash256(j)
    str := hex.EncodeToString(h[len(h) - AddressLen:])
    return Address(str)
}
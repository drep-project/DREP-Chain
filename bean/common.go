package bean

import (
    "BlockChainTest/mycrypto"
    "encoding/hex"
)

const (
    MsgTypeBlockHeader = iota
    MsgTypeBlock
    MsgTypeTransaction
    MsgTypeSetUp
    MsgTypeCommitment
    MsgTypeChallenge
    MsgTypeResponse
    MsgTypeNewPeer
    MsgTypePeerList
    MsgTypeBlockReq
    MsgTypeBlockResp
    MsgTypePing
    MsgTypePong
    MsgTypeOfflinePeers
    MsgTypeFirstPeerInfoList
    MsgTypeAccount
    MsgTypeLog
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
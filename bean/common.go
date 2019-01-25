package bean

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
)

type Address string

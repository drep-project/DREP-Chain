package bean

type Address string

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
)
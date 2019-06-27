package types

//本模块的消息只能在调用本模块（consensus及对应的子模块）的函数中使用，否则会出错
//例如MsgTypeCommitment消息，在consensus中定义的，发送和接收此消息必须使用consensus中的函数
const (
	MsgTypeSetUp      = 0
	MsgTypeCommitment = 1
	MsgTypeResponse   = 2
	MsgTypeChallenge  = 3
	MsgTypeFail       = 4

	MaxMsgSize = 20 << 20
)

var NumberOfMsg = 5

type RouteMsgWrap struct {
	Peer     *PeerInfo
	SetUpMsg *Setup
}

//
//type Setup struct {
//	Height uint64
//	Msg    []byte
//}
//
//type Commitment struct {
//	Height uint64
//	Q      *secp256k1.PublicKey
//}
//
//type Challenge struct {
//	Height      uint64
//	SigmaPubKey *secp256k1.PublicKey
//	SigmaQ      *secp256k1.PublicKey
//	R           []byte
//}
//
//type Response struct {
//	Height uint64
//	S      []byte
//}
//
//type Fail struct {
//	Height uint64
//	Reason string
//}

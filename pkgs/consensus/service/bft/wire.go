package bft

import (
	"encoding/json"
	"github.com/drep-project/binary"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/pkgs/consensus/types"
)

//本模块的消息只能在调用本模块（consensus及对应的子模块）的函数中使用，否则会出错
//例如MsgTypeCommitment消息，在consensus中定义的，发送和接收此消息必须使用consensus中的函数
const (
	MsgTypeSetUp      = 0
	MsgTypeCommitment = 1
	MsgTypeResponse   = 2
	MsgTypeChallenge  = 3
	MsgTypeFail       = 4
	MsgTypeValidateReq = 5
	MsgTypeValidateRes = 6

	MaxMsgSize = 20 << 20
)

var NumberOfMsg = 8

type MsgWrap struct {
	Peer types.IPeerInfo
	Code uint64
	Msg  []byte
}

type Setup struct {
	Height uint64

	Msg []byte
}

func (setup *Setup) String() string {
	bytes, _ := json.Marshal(setup)
	return string(bytes)
}

type Commitment struct {
	Height uint64
	BpKey  *secp256k1.PublicKey
	Q      *secp256k1.PublicKey
}

func (commitment *Commitment) String() string {
	bytes, _ := json.Marshal(commitment)
	return string(bytes)
}

type Challenge struct {
	Height uint64

	SigmaPubKey *secp256k1.PublicKey
	SigmaQ      *secp256k1.PublicKey
	R           []byte
}

func (Challenge *Challenge) String() string {
	bytes, _ := json.Marshal(Challenge)
	return string(bytes)
}

type Response struct {
	Height uint64
	BpKey  *secp256k1.PublicKey
	S      []byte
}

func (response *Response) String() string {
	bytes, _ := json.Marshal(response)
	return string(bytes)
}

type Fail struct {
	Height uint64

	Reason string
}

func (fail *Fail) String() string {
	bytes, _ := json.Marshal(fail)
	return string(bytes)
}

type IConsenMsg interface {
	AsSignMessage() []byte
	AsMessage() []byte
}

type CompletedBlockMessage struct {
	MultiSignature
	StateRoot []byte //sencond round  leader should send stateroot, then member verify
}

func (completedBlockMessage *CompletedBlockMessage) AsSignMessage() []byte {
	bytes, _ := binary.Marshal(completedBlockMessage)
	return bytes
}

func (completedBlockMessage *CompletedBlockMessage) AsMessage() []byte {
	return completedBlockMessage.AsSignMessage()
}

func CompletedBlockFromMessage(bytes []byte) (*CompletedBlockMessage, error) {
	completedBlockMessage := &CompletedBlockMessage{}
	err := binary.Unmarshal(bytes, completedBlockMessage)
	if err != nil {
		return nil, err
	}
	return completedBlockMessage, nil
}

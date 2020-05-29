package bft

import (
	"encoding/json"
	"fmt"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/pkgs/consensus/types"
	"github.com/drep-project/binary"
)

//The messages of this module can only be used in functions that call this module (consensus and its corresponding submodules)
//For example, the MsgTypeCommitment message, defined in consensus, must be sent and received using a function in consensus
const (
	MsgTypeSetUp       = 0
	MsgTypeCommitment  = 1
	MsgTypeResponse    = 2
	MsgTypeChallenge   = 3
	MsgTypeFail        = 4
	MsgTypeValidateReq = 5
	MsgTypeValidateRes = 6

	MaxMsgSize = 20 << 20

	SetupMagic    = 0xfefefbfe
	CommitMagic   = 0xfefefbfd
	ChallegeMagic = 0xfefefbfc
	FailMagic     = 0xfefefbfb
	ResponseMagic = 0xfefefbfa
	//ValidateReqMagic = 0xfefefbf9
	//validateResMagic = 0xfefefbf8
)

var NumberOfMsg = 7

type MsgWrap struct {
	Peer types.IPeerInfo
	Code uint64
	Msg  []byte
}

type Setup struct {
	Height uint64
	Magic  uint32
	Round  int

	Msg []byte
}

func (setup *Setup) String() string {
	bytes, _ := json.Marshal(setup)
	return string(bytes)
}

type Commitment struct {
	Height uint64
	Magic  uint32
	Round  int
	BpKey  *secp256k1.PublicKey
	Q      *secp256k1.PublicKey
}

func (commitment *Commitment) String() string {
	bytes, _ := json.Marshal(commitment)
	return string(bytes)
}

type Challenge struct {
	Height      uint64
	Magic       uint32
	Round       int
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
	Magic  uint32
	Round  int
	BpKey  *secp256k1.PublicKey
	S      []byte
}

func (response *Response) String() string {
	bytes, _ := json.Marshal(response)
	return string(bytes)
}

type Fail struct {
	Height uint64
	Magic  uint32
	Round  int

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
	if completedBlockMessage == nil {
		return nil, fmt.Errorf("CompletedBlockFromMessage err")
	}
	return completedBlockMessage, nil
}

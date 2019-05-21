package types

import (
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
)

type Setup struct {
	Height uint64

	Msg                  []byte
}

type Commitment struct {
	Height uint64
	BpKey                    *secp256k1.PublicKey
	Q                    *secp256k1.PublicKey
}

type Challenge struct {
	Height uint64

	SigmaPubKey          *secp256k1.PublicKey
	SigmaQ               *secp256k1.PublicKey
	R                    []byte
}

type Response struct {
	Height uint64
	BpKey                 *secp256k1.PublicKey
	S                     []byte
}

type Fail struct {
	Height uint64

	Reason string
}

type IConsenMsg interface {
	AsSignMessage() []byte
	AsMessage() []byte
}

type ResponseWiteRootMessage struct {
	types.MultiSignature
	StateRoot			 []byte  //sencond round  leader should send stateroot, then member verify
}

func (responseWiteRootMessage *ResponseWiteRootMessage) AsSignMessage() []byte {
	bytes, _ := binary.Marshal(responseWiteRootMessage)
	return bytes
}

func (responseWiteRootMessage *ResponseWiteRootMessage) AsMessage() []byte {
	return responseWiteRootMessage.AsSignMessage()
}

func ResponseWiteRootFromMessage(bytes []byte) (*ResponseWiteRootMessage, error) {
	responseWiteRootMessage := &ResponseWiteRootMessage{}
	err := binary.Unmarshal(bytes, responseWiteRootMessage)
	if err != nil {
		return nil, err
	}
	return responseWiteRootMessage, nil
}
package bft

import (
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/network/p2p"
)

type Sender interface {
	SendAsync(w p2p.MsgWriter, msgType uint64, msg interface{}) chan error
}

type MultiSignature struct {
	Sig    secp256k1.Signature
	Leader int
	Bitmap []byte
}

func newMultiSignature(sig secp256k1.Signature, leader int, bitmap []byte) *MultiSignature {
	return &MultiSignature{sig, leader, bitmap}
}

func (multiSignature *MultiSignature) AsSignMessage() []byte {
	bytes, _ := binary.Marshal(multiSignature)
	return bytes
}

func (multiSignature *MultiSignature) AsMessage() []byte {
	return multiSignature.AsSignMessage()
}

func (multiSignature *MultiSignature) Num() int {
	num := 0
	for _, val := range multiSignature.Bitmap {
		if val == 1 {
			num++
		}
	}
	return num
}

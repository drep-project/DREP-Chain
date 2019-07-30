package bft

import (
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
)

type MultiSignature struct {
	Sig    secp256k1.Signature
	Bitmap []byte
}

func (multiSignature *MultiSignature) AsSignMessage() []byte {
	bytes, _ := binary.Marshal(multiSignature)
	return bytes
}

func (multiSignature *MultiSignature) AsMessage() []byte {
	return multiSignature.AsSignMessage()
}

func MultiSignatureFromMessage(bytes []byte) (*MultiSignature, error) {
	multySig := &MultiSignature{}
	err := binary.Unmarshal(bytes, multySig)
	if err != nil {
		return nil, err
	}
	return multySig, nil
}

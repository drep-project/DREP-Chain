package types

import (
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
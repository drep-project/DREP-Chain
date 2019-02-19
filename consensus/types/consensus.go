package types

import (
	"github.com/drep-project/drep-chain/crypto/secp256k1"
)

type Setup struct {
	Height int64

	Msg                  []byte
}

type Commitment struct {
	Height int64

	Q                    *secp256k1.PublicKey
}

type Challenge struct {
	Height int64

	SigmaPubKey          *secp256k1.PublicKey
	SigmaQ               *secp256k1.PublicKey
	R                    []byte
}

type Response struct {
	Height int64

	S                     []byte
}

type Fail struct {
	Height int64

	Reason string
}
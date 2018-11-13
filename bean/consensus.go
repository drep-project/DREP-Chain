package bean

import "BlockChainTest/mycrypto"

type Setup struct {
	Msg                  []byte
	PubKey               *mycrypto.Point
	Sig                  *mycrypto.Signature
}

type Commitment struct {
	PubKey               *mycrypto.Point
	Q                    *mycrypto.Point
}

type Challenge struct {
	SigmaPubKey          *mycrypto.Point
	SigmaQ               *mycrypto.Point
	R                    []byte
}

type Response struct {
	PubKey               *mycrypto.Point
	S                    []byte
}
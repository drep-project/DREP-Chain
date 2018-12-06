package bean

import "BlockChainTest/mycrypto"

type Setup struct {
	Msg                  []byte
}

type Commitment struct {
	Q                    *mycrypto.Point
}

type Challenge struct {
	SigmaPubKey          *mycrypto.Point
	SigmaQ               *mycrypto.Point
	R                    []byte
}

type Response struct {
	S                    []byte
}
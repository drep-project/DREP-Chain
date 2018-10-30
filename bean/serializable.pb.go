package bean

import "BlockChainTest/mycrypto"

type Serializable struct {
	Header               int32
	Body                 []byte
	PubKey               *mycrypto.Point
	Sig                  *mycrypto.Signature
}
package types

import (
	"github.com/drep-project/drep-chain/crypto/secp256k1"
)

type ConsensusConfig struct {
	ConsensusMode string `json:"consensusMode"`
	//Producers       Producers            `json:"producers"` // key对应的是ip，value 对应的secp256k1.PublicKey
	MyPk   *secp256k1.PublicKey `json:"mypk"`
	Enable bool                 `json:"enable"`
}

//TODO how to identify a mine pk or pr&addr
//type Producers map[string]*secp256k1.PublicKey
//
//func  NewProducers()Producers{
//	return make(Producers)
//}

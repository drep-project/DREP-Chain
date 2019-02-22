package types

import (
	"github.com/drep-project/drep-chain/crypto/secp256k1"
)

type ConsensusConfig struct {
	ConsensusMode string          `json:"consensusMode"`
	Producers []*Producer         `json:"producers"`
	MyPk 	 *secp256k1.PublicKey `json:"mypk"`
}

//TODO how to identify a mine pk or pr&addr
type Producer struct {
	Public  *secp256k1.PublicKey
	Ip string
	Port int
}
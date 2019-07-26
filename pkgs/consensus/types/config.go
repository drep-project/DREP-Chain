package types

import (
	"github.com/drep-project/drep-chain/crypto/secp256k1"
)

type ConsensusConfig struct {
	ConsensusMode string `json:"consensusMode"`
	MyPk   *secp256k1.PublicKey `json:"mypk"`
	Enable bool                 `json:"enable"`
	Producers        []Producer     `json:"producers"`
}

type Producer struct {
	Pubkey *secp256k1.PublicKey `json:"pubkey"`
	IP     string               `json:"ip"`
}

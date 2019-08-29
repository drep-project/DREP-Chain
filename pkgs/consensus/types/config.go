package types

import (
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
)

const (
	Solo = iota
	Pbft
)

type ConsensusConfig struct {
	ConsensusMode string               `json:"consensusMode"`
	MyPk          *secp256k1.PublicKey `json:"mypk"`
	Enable        bool                 `json:"enable"`
	Producers     ProducerSet          `json:"producers"`
}

type Producer struct {
	Pubkey *secp256k1.PublicKey `json:"pubkey"`
	IP     string               `json:"ip"`
}

func (producer *Producer) Address() crypto.CommonAddress {
	return crypto.PubkeyToAddress(producer.Pubkey)
}

type ProducerSet []Producer

func (produceSet *ProducerSet) IsLocalIP(ip string) bool {
	for _, bp := range *produceSet {
		if bp.IP == ip {
			return true
		}
	}
	return false
}

func (produceSet *ProducerSet) IsLocalPk(pk *secp256k1.PublicKey) bool {
	for _, bp := range *produceSet {
		if bp.Pubkey.IsEqual(pk) {
			return true
		}
	}
	return false
}

func (produceSet *ProducerSet) IsLocalAddress(addr crypto.CommonAddress) bool {
	for _, bp := range *produceSet {
		if crypto.PubkeyToAddress(bp.Pubkey) == addr {
			return true
		}
	}
	return false
}

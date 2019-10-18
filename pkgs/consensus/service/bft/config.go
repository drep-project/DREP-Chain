package bft

import (
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/network/p2p/enode"
)

type BftConfig struct {
	MyPk          *secp256k1.PublicKey `json:"mypk"`
	StartMiner    bool                 `json:"startMiner"`
	ProducerNum  int	`json:"producerNum"`
	BlockInterval int                  `json:"blockInterval"`
	ChangeInterval int                  `json:"changeInterval"`
}

type Producer struct {
	Pubkey *secp256k1.PublicKey `json:"pubkey"`
	Node   *enode.Node
}

func (producer *Producer) Address() crypto.CommonAddress {
	return crypto.PubkeyToAddress(producer.Pubkey)
}

type ProducerSet []*Producer

func (produceSet *ProducerSet) IsLocalIP(ip string) bool {
	for _, bp := range *produceSet {
		if bp.Node.IP().String() == ip {
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
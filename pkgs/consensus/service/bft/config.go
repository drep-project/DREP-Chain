package bft

import (
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
)

type BftConfig struct {
	MyPk           *secp256k1.PublicKey `json:"mypk"`
	StartMiner     bool                 `json:"startMiner"`
	ProducerNum    int                  `json:"producerNum"`
	BlockInterval  int64                `json:"blockInterval"`
	ChangeInterval uint64               `json:"changeInterval"`
}

package solo

import "github.com/drep-project/DREP-Chain/crypto/secp256k1"

type SoloConfig struct {
	MyPk           *secp256k1.PublicKey `json:"mypk"`
	StartMiner     bool                 `json:"startMiner"`
	BlockInterval  int                  `json:"blockInterval"`
	ChangeInterval uint64               `json:"changeInterval"`
}

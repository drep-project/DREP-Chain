package solo

import "github.com/drep-project/drep-chain/crypto/secp256k1"

type SoloConfig struct {
	MyPk          *secp256k1.PublicKey `json:"mypk"`
	Miner     bool                 `json:"miner"`
}
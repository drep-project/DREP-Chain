package types

import (
	"github.com/drep-project/drep-chain/crypto/secp256k1"
)

type P2pConfig struct {
	PrvKey *secp256k1.PrivateKey  `json:"omitempty"`
	ListerAddr string               `json:"omitempty"`
	Port int
	BootNodes []BootNode  //pub@Addr
}

type BootNode struct {
	IP string                       `json:"ip"`
	Port int                        `json:"port"`
}

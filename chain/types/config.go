package types

import "github.com/drep-project/drep-chain/common"

type ChainConfig struct {
	RemotePort int 				`json:"remoteport"`
	ChainId common.ChainIdType 	`json:"chainId"`
	GenesisPK  string `json:"genesispk"`
}
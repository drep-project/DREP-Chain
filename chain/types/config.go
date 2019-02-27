package types

import "github.com/drep-project/drep-chain/app"

type ChainConfig struct {
	RemotePort int 				`json:"remoteport"`
	RootChain app.ChainIdType	`json:"rootChain,omitempty"`
	ChainId app.ChainIdType 	`json:"chainId"`
	GenesisPK  string `json:"genesispk"`
}
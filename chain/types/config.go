package types

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
)

type ChainConfig struct {
	RemotePort 			int 				`json:"remoteport"`
	RootChain 			app.ChainIdType		`json:"rootChain,omitempty"`
	ChainId 			app.ChainIdType 	`json:"chainId"`
	Producers 			[]*Producer         `json:"producers"`
	GenesisPK  			string 				`json:"genesispk"`
	SkipCheckMutiSig  	bool 				`json:"skipCheckMutiSig"`
}

//TODO how to identify a mine pk or pr&addr
type Producer struct {
	Public  	*secp256k1.PublicKey
	Ip 			string
	Port 		int
}
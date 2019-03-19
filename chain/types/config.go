package types

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
)

type ChainConfig struct {
	RemotePort 	int 			`json:"remoteport"`
	RootChain 	app.ChainIdType	`json:"rootChain,omitempty"`
	ChainId 	app.ChainIdType `json:"chainId"`
	Bios  		string 			`json:"bios"`
	Producers 	[]Producer      `json:"producers"`
}
//TODO how to identify a mine pk or pr&addr
type Producer struct {
	Account  	string
	Pubkey 		secp256k1.PublicKey
	ChainCode 	[]byte
	SignPubkey 	secp256k1.PublicKey
	Ip 			string
	Port 		int
}
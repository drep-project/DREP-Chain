package types

import "github.com/drep-project/drep-chain/crypto/secp256k1"

type ConsensusConfig struct {
	ConsensusMode 	string          	`json:"consensusMode"`
	Producers 		[]*Producer         `json:"producers"`
	Me 	 			string 				`json:"me"`
	EnableConsensus bool				`json:"enableConsensus"`
}

//TODO how to identify a mine pk or pr&addr
type Producer struct {
	Account  	string
	SignPubkey 	secp256k1.PublicKey
	Ip 			string
	Port 		int
}
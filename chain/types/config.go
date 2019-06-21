package types

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
)

type ChainConfig struct {
	RemotePort       int             `json:"remoteport"`
	RootChain        app.ChainIdType `json:"rootChain,omitempty"`
	ChainId          app.ChainIdType `json:"chainId,omitempty"`
	GenesisPK        string          `json:"genesispk"`
	SkipCheckMutiSig bool            `json:"skipCheckMutiSig"`
	Producers        []Producers     `json:"producers"`

}

type BlockMgrConfig struct {
	GasPrice		 OracleConfig    `json:"gasprice"`
	JournalFile      string          `json:"journalFile"`
}

type Producers struct {
	Pubkey *secp256k1.PublicKey 	`json:"pubkey"`
	IP     string               	`json:"ip"`
}

type OracleConfig struct {
	Blocks     int					`json:"blocks"`
	Percentile int					`json:"percentile"`
	Default    uint64				`json:"default"`
	MaxPrice   uint64				`json:"maxPrice"`
}
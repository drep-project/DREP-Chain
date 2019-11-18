package chain

import (
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/types"
)

type ChainConfig struct {
	RemotePort  int                  `json:"remoteport"`
	RootChain   types.ChainIdType    `json:"rootChain,omitempty"`
	ChainId     types.ChainIdType    `json:"chainId,omitempty"`
	GenesisAddr crypto.CommonAddress `json:"genesisaddr"`
}

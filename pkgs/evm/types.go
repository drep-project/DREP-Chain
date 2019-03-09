package evm

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/crypto"
	"math/big"
)

type VMConfig struct {

}

type Message struct {
	From      crypto.CommonAddress
	To        crypto.CommonAddress
	ChainId   app.ChainIdType
	DestChain app.ChainIdType
	Gas       uint64
	Value     *big.Int
	Nonce     uint64
	Input     []byte
	ReadOnly  bool
}
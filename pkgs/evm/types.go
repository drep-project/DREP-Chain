package evm

import (
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/app"
	"math/big"
)

type VMConfig struct {

}

type Message struct {
	From      string
	ChainId   app.ChainIdType
	DestChain app.ChainIdType
	Gas       uint64
	Value     *big.Int
	Nonce     uint64
	Input     []byte
	ReadOnly  bool
	MessageType chainTypes.TxType
}
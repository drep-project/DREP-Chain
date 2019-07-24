package types

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/crypto"
)

type Log struct {
	Address crypto.CommonAddress
	ChainId app.ChainIdType
	TxHash  crypto.Hash
	Topics  []crypto.Hash
	Data    []byte
	Height  int64
	Removed bool
}

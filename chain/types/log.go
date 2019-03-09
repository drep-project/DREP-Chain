package types

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/crypto"
)

type Log struct {
	Address crypto.CommonAddress
	ChainId app.ChainIdType
	TxHash  []byte
	Topics  [][]byte
	Data    []byte
}
package types

import (
	"github.com/drep-project/drep-chain/app"
)

type Log struct {
	Name 	string
	ChainId app.ChainIdType
	TxHash  []byte
	Topics  [][]byte
	Data    []byte
}
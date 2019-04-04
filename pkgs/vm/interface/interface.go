package _interface

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/chain/types"
)

type Vm interface {
	app.Service
	ApplyMessage(message *types.Message) ([]byte, uint64, error)
}
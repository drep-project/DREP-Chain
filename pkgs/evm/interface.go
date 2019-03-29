package evm

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/chain/types"
)

type Vm interface {
	app.Service
	ApplyMessage(message *types.Message) (uint64, error)
}
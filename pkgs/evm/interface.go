package evm

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/pkgs/evm/vm"
)

type Vm interface {
	app.Service
	ApplyMessage(evm *vm.EVM, message *types.Message) (uint64, error)
}
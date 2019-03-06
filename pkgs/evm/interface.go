package evm

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/pkgs/evm/vm"
	"github.com/drep-project/drep-chain/transaction/types"
)

type Vm interface {
	app.Service
	ApplyTransaction(evm *vm.EVM, tx *types.Transaction) (uint64, error)
}
package evm

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/pkgs/evm/vm"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
)

type Vm interface {
	app.Service
	ApplyTransaction(evm *vm.EVM, tx *chainTypes.Transaction) (uint64, error)
}
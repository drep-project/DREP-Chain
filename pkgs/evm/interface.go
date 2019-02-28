package evm

import (
	"github.com/drep-project/drep-chain/app"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/pkgs/evm/vm"
)

type Vm interface {
	app.Service
	ApplyTransaction(evm *vm.EVM, tx *chainTypes.Transaction) (uint64, error)
}
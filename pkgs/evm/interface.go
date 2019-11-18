package evm

import (
	"github.com/drep-project/DREP-Chain/app"
	"github.com/drep-project/DREP-Chain/pkgs/evm/vm"
	"github.com/drep-project/DREP-Chain/types"
	"math/big"
)

type Vm interface {
	app.Service
	Eval(vm.VMState, *types.Transaction, *types.BlockHeader, ChainContext, uint64, *big.Int) (ret []byte, gasUsed uint64, failed bool, err error)
}

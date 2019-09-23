package evm

import (
	"github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/params"
	"github.com/drep-project/drep-chain/pkgs/evm/vm"
	"github.com/drep-project/drep-chain/types"
)

var (
	_ = (chain.ITransactionSelector)((*EvmDeployTransactionSelector)(nil))
	//_ = (chain.ITransactionSelector)((*EvmCallTransactionSelector)(nil))

	_ = (chain.ITransactionValidator)((*EvmDeployTransactionExecutor)(nil))
	//_ = (chain.ITransactionValidator)((*EvmCallTransactionExecutor)(nil))
)

// ***********DEPLOY**************//
type EvmDeployTransactionSelector struct{}

func (evmDeployTransactionSelector *EvmDeployTransactionSelector) Select(tx *types.Transaction) bool {
	return tx.Type() == types.CreateContractType || tx.Type() == types.CallContractType
}

type EvmDeployTransactionExecutor struct {
	vm *EvmService
}

func (vmDeployTransactionExecutor *EvmDeployTransactionExecutor) ExecuteTransaction(context *chain.ExecuteTransactionContext) ([]byte, bool, []*types.Log, error) {
	tx := context.Tx()
	contractCreation := tx.To() == nil || tx.To().IsEmpty()
	// Set the starting gas for the raw transaction
	gas := uint64(0)
	if contractCreation {
		gas = vm.TxGasContractCreation
	} else {
		gas = params.TxGas
	}
	err :=  context.UseGas(gas)
	if err != nil {
		return nil, false, nil, err
	}
	state := vm.NewState(context.TrieStore())
	ret, gas, failed, err := vmDeployTransactionExecutor.vm.Eval(
		state,
		context.Tx(),
		context.Header(),
		vmDeployTransactionExecutor.vm.Chain,
		context.GasRemained(),
		context.Value())
	context.UseGas(gas)

	refund := context.GasUsed() / 2
	if refund > state.GetRefund() {
		refund = state.GetRefund()
	}
	context.RefundGas(refund)
	return ret, failed, state.GetLogs(context.Tx().TxHash()), err
}

// ***********CALL**************//
/*type EvmCallTransactionSelector struct {}
func (evmCallTransactionSelector *EvmCallTransactionSelector) Select(tx *types.Transaction) bool {
	return tx.Type() == types.CallContractType
}

type EvmCallTransactionExecutor struct {

}

func (evmCallTransactionExecutor *EvmCallTransactionExecutor) ExecuteTransaction(context *chain.ExecuteTransactionContext) ([]byte, bool, []*types.Log, error) {
	panic("not imple")
}*/

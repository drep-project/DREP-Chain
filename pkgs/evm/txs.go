package evm

import (
	"github.com/drep-project/drep-chain/chain"
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
	state := vm.NewState(context.TrieStore(), context.Header().Height)

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

	logs := state.GetLogs(context.Tx().TxHash())
	for _,log := range logs {
		log.TxType = context.Tx().Type()
	}
	return ret, failed, logs, err
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

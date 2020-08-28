package transactions

import (
	"encoding/json"

	"github.com/drep-project/DREP-Chain/pkgs/evm"
	"github.com/drep-project/DREP-Chain/pkgs/evm/vm"

	"github.com/drep-project/DREP-Chain/params"

	"github.com/drep-project/DREP-Chain/types"
)

//var (
//	_ = (ITransactionValidator)((*Processor)(nil))
//)

type Processor struct {
}

func (processor *Processor) subBalance(context *ExecuteTransactionContext) error {
	originBalance := context.TrieStore().GetBalance(context.From(), context.Header().Height)
	leftBalance := originBalance.Sub(originBalance, context.Tx().Amount())
	if leftBalance.Sign() < 0 {
		return ErrBalance
	}
	err := context.TrieStore().PutBalance(context.From(), context.Header().Height, leftBalance)
	if err != nil {
		return err
	}
	return nil
}

func (processor *Processor) ExecuteTransaction(context *ExecuteTransactionContext) *types.ExecuteTransactionResult {
	etr := &types.ExecuteTransactionResult{}
	detail := &types.CancelCreditDetail{}
	store := context.TrieStore() // GetBalance
	height := context.Header().Height
	from := context.From()
	tx := context.Tx()
	to := tx.To()
	amount := tx.Amount()
	data := tx.GetData()
	txType := tx.Type()
	var err error

	switch txType {
	case types.TransferType:
		err = processor.subBalance(context)
		if err != nil {
			etr.Txerror = err
			return etr
		}
		toBalance := store.GetBalance(to, height)
		addBalance := toBalance.Add(toBalance, amount)
		err = store.PutBalance(to, height, addBalance)
	case types.VoteCreditType:
		err = processor.subBalance(context)
		if err != nil {
			etr.Txerror = err
			return etr
		}
		err = store.VoteCredit(from, to, amount, height)
	case types.CandidateType:
		err = processor.subBalance(context)
		if err != nil {
			etr.Txerror = err
			return etr
		}
		err = store.CandidateCredit(from, amount, data, height)
	case types.CancelVoteCreditType:
		detail, err = store.CancelVoteCredit(from, to, amount, height)
	case types.CancelCandidateType:
		detail, err = store.CancelCandidateCredit(from, amount, height)
	case types.CreateContractType, types.CallContractType:
		evmService := &evm.EvmService{}
		evmService.Config = evm.DefaultEvmConfig
		state := vm.NewState(store, height)
		ret, gas, addr, failed, err := evmService.Eval(state, tx, context.Header(), context.GasRemained(), context.Value())
		if err != nil {
			etr.Txerror = err
			return etr
		}
		etr.TxResult = ret
		etr.ContractAddr = addr
		etr.ContractTxExecuteFail = failed

		err = context.UseGas(gas)
		refund := context.GasUsed() / 2
		if refund > state.GetRefund() {
			refund = state.GetRefund()
		}
		context.RefundGas(refund)
	case types.SetAliasType:
		alias := data
		err = store.AliasSet(from, string(alias), height)
		if err != nil {
			etr.Txerror = err
			return etr
		}
		err = context.UseGas(params.AliasGas * uint64(len(alias)))
	default:
		err = ErrTxUnSupport
	}

	if err != nil {
		etr.Txerror = err
		return etr
	}

	if detail != nil {
		logs := make([]*types.Log, 0, 1)
		data, _ := json.Marshal(detail)
		log := types.Log{TxType: tx.Type(), TxHash: *tx.TxHash(), Data: data, Height: height, TxIndex: 0}
		logs = append(logs, &log)
	}

	err = store.PutNonce(from, tx.Nonce()+1)
	if err != nil {
		etr.Txerror = err
		return etr
	}
	return etr
}

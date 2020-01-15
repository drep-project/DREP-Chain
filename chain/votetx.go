package chain

import (
	"github.com/drep-project/DREP-Chain/types"
)

/**********************stake********************/

type StakeTxSelector struct {
}

func (StakeTxSelector *StakeTxSelector) Select(tx *types.Transaction) bool {
	return tx.Type() == types.VoteCreditType
}

var (
	_ = (ITransactionSelector)((*StakeTxSelector)(nil))
	_ = (ITransactionValidator)((*StakeTransactionProcessor)(nil))
)

type StakeTransactionProcessor struct {
}

func (processor *StakeTransactionProcessor) ExecuteTransaction(context *ExecuteTransactionContext) *types.ExecuteTransactionResult {
	etr := &types.ExecuteTransactionResult{}
	from := context.From()
	ts := context.TrieStore()
	tx := context.Tx()

	originBalance := ts.GetBalance(from, context.header.Height)
	leftBalance := originBalance.Sub(originBalance, tx.Amount())
	if leftBalance.Sign() < 0 {
		etr.Txerror = ErrBalance
		return etr
	}

	err := ts.PutBalance(from, context.header.Height, leftBalance)
	if err != nil {
		etr.Txerror = err
		return etr
	}

	err = ts.VoteCredit(from, tx.To(), tx.Amount(), context.header.Height)
	if err != nil {
		etr.Txerror = err
		return etr
	}
	err = ts.PutNonce(from, tx.Nonce()+1)
	if err != nil {
		etr.Txerror = err
		return etr
	}

	return etr
}

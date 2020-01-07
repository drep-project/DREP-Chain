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

func (processor *StakeTransactionProcessor) ExecuteTransaction(context *ExecuteTransactionContext) ([]byte, bool, []*types.Log, error) {
	from := context.From()
	ts := context.TrieStore()
	tx := context.Tx()

	originBalance := ts.GetBalance(from, context.header.Height)
	leftBalance := originBalance.Sub(originBalance, tx.Amount())
	if leftBalance.Sign() < 0 {
		return nil, false, nil, ErrBalance
	}

	err := ts.PutBalance(from, context.header.Height, leftBalance)
	if err != nil {
		return nil, false, nil, err
	}

	err = ts.VoteCredit(from, tx.To(), tx.Amount(), context.header.Height)
	if err != nil {
		return nil, false, nil, err
	}
	err = ts.PutNonce(from, tx.Nonce()+1)
	if err != nil {
		return nil, false, nil, err
	}

	return nil, true, nil, err
}

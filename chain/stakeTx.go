package chain

import (
	"github.com/drep-project/drep-chain/types"
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
	store := context.TrieStore()
	tx := context.Tx()
	stakeStore := context.TrieStakeStore()

	originBalance := store.GetBalance(from)
	toBalance := store.GetBalance(tx.To())
	leftBalance := originBalance.Sub(originBalance, tx.Amount())
	if leftBalance.Sign() < 0 {
		return nil, false, nil, ErrBalance
	}
	addBalance := toBalance.Add(toBalance, tx.Amount())
	err := store.PutBalance(from, leftBalance)
	if err != nil {
		return nil, false, nil, err
	}
	err = stakeStore.VoteCredit(from, tx.To(), addBalance)
	if err != nil {
		return nil, false, nil, err
	}
	err = store.PutNonce(from, tx.Nonce()+1)
	if err != nil {
		return nil, false, nil, err
	}

	return nil, true, nil, err
}

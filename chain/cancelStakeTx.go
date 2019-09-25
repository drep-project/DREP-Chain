package chain

import (
"github.com/drep-project/drep-chain/types"
)

/**********************stake********************/

type CancelStakeTxSelector struct {
}

func (CancelStakeTxSelector *CancelStakeTxSelector) Select(tx *types.Transaction) bool {
	return tx.Type() == types.CancelVoteCreditType
}

var (
	_ = (ITransactionSelector)((*CancelStakeTxSelector)(nil))
	_ = (ITransactionValidator)((*CancelStakeTransactionProcessor)(nil))
)

type CancelStakeTransactionProcessor struct {
}

func (processor *CancelStakeTransactionProcessor) ExecuteTransaction(context *ExecuteTransactionContext) ([]byte, bool, []*types.Log, error) {
	from := context.From()
	tx := context.Tx()
	stakeStore := context.TrieStore()

	err := stakeStore.CancelVoteCredit(from, tx.To(), tx.Amount(), context.blockContext.Block.Header.Height)
	if err != nil {
		return nil, false, nil, err
	}

	return nil, true, nil, err
}

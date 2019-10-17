package chain

import (
	"github.com/drep-project/drep-chain/types"
)

/**********************register********************/

type RegCandidateTxSelector struct {
}

func (StakeTxSelector *RegCandidateTxSelector) Select(tx *types.Transaction) bool {
	return tx.Type() == types.CandidateType
}

var (
	_ = (ITransactionSelector)((*StakeTxSelector)(nil))
	_ = (ITransactionValidator)((*CandidateTransactionProcessor)(nil))
)

type CandidateTransactionProcessor struct {
}

func (processor *CandidateTransactionProcessor) ExecuteTransaction(context *ExecuteTransactionContext) ([]byte, bool, []*types.Log, error) {
	from := context.From()
	store := context.TrieStore()
	tx := context.Tx()

	originBalance := store.GetBalance(from, context.header.Height)
	if originBalance.Cmp(tx.Amount()) < 0 {
		return nil, false, nil, ErrBalance
	}
	leftBalance := originBalance.Sub(originBalance, tx.Amount())
	if leftBalance.Sign() < 0 {
		return nil, false, nil, ErrBalance
	}

	cd := types.CandidateData{}
	if err := cd.Unmarshal(tx.GetData()); nil != err {
		return nil, false, nil, err
	}
	err := store.CandidateCredit(from, tx.Amount(), tx.GetData())
	if err != nil {
		return nil, false, nil, err
	}
	err = store.PutNonce(from, tx.Nonce()+1)
	if err != nil {
		return nil, false, nil, err
	}

	return nil, true, nil, err
}

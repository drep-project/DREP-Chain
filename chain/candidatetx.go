package chain

import (
	"github.com/drep-project/DREP-Chain/types"
)

/**********************register********************/

type CandidateTxSelector struct {
}

func (StakeTxSelector *CandidateTxSelector) Select(tx *types.Transaction) bool {
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
		log.WithField("originBalance", originBalance).WithField("tx amount", tx.Amount()).Info("no enough balance for candidate")
		return nil, false, nil, ErrBalance
	}
	leftBalance := originBalance.Sub(originBalance, tx.Amount())
	if leftBalance.Sign() < 0 {
		return nil, false, nil, ErrBalance
	}

	err := store.PutBalance(from, context.header.Height, leftBalance)
	if err != nil {
		return nil, false, nil, err
	}

	cd := types.CandidateData{}
	if err := cd.Unmarshal(tx.GetData()); nil != err {
		return nil, false, nil, err
	}
	err = store.CandidateCredit(from, tx.Amount(), tx.GetData(), context.header.Height)
	if err != nil {
		return nil, false, nil, err
	}
	err = store.PutNonce(from, tx.Nonce()+1)
	if err != nil {
		return nil, false, nil, err
	}

	return nil, true, nil, err
}

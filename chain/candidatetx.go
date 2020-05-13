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

func (processor *CandidateTransactionProcessor) ExecuteTransaction(context *ExecuteTransactionContext) *types.ExecuteTransactionResult {
	etr := &types.ExecuteTransactionResult{}
	from := context.From()
	store := context.TrieStore()
	tx := context.Tx()

	originBalance := store.GetBalance(from, context.header.Height)
	if originBalance.Cmp(tx.Amount()) < 0 {
		log.WithField("originBalance", originBalance).WithField("tx amount", tx.Amount()).Info("no enough balance for candidate")
		etr.Txerror = ErrBalance
		return etr
	}
	leftBalance := originBalance.Sub(originBalance, tx.Amount())
	if leftBalance.Sign() < 0 {
		etr.Txerror = ErrBalance
		return etr
	}

	err := store.PutBalance(from, context.header.Height, leftBalance)
	if err != nil {
		etr.Txerror = err
		return etr
	}

	cd := types.CandidateData{}
	if err := cd.Unmarshal(tx.GetData()); nil != err {
		etr.Txerror = err
		return etr
	}
	err = store.CandidateCredit(from, tx.Amount(), tx.GetData(), context.header.Height)
	if err != nil {
		etr.Txerror = err
		return etr
	}
	err = store.PutNonce(from, tx.Nonce()+1)
	if err != nil {
		etr.Txerror = err
		return etr
	}

	return etr
}

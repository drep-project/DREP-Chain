package chain

import "github.com/drep-project/DREP-Chain/types"

/****************** transfer ********************/
type TransferTxSelector struct {
}

func (transferTxSelector *TransferTxSelector) Select(tx *types.Transaction) bool {
	return tx.Type() == types.TransferType
}

var (
	_ = (ITransactionValidator)((*TransferTransactionProcessor)(nil))
)

type TransferTransactionProcessor struct {
}

func (transferTransactionProcessor *TransferTransactionProcessor) ExecuteTransaction(context *ExecuteTransactionContext) ([]byte, bool, []*types.Log, error) {
	from := context.From()
	store := context.TrieStore() // GetBalance
	tx := context.Tx()
	originBalance := store.GetBalance(from, context.header.Height)
	toBalance := store.GetBalance(tx.To(), context.header.Height)
	leftBalance := originBalance.Sub(originBalance, tx.Amount())
	if leftBalance.Sign() < 0 {
		return nil, false, nil, ErrBalance
	}
	addBalance := toBalance.Add(toBalance, tx.Amount())
	err := store.PutBalance(from, context.header.Height, leftBalance)
	if err != nil {
		return nil, false, nil, err
	}

	err = store.PutBalance(tx.To(), context.header.Height, addBalance)
	if err != nil {
		return nil, false, nil, err
	}

	err = store.PutNonce(from, tx.Nonce()+1)
	if err != nil {
		return nil, false, nil, err
	}
	return nil, true, nil, nil
}

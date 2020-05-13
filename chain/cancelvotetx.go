package chain

import (
	"encoding/json"
	"github.com/drep-project/DREP-Chain/types"
)

/**********************stake********************/

type CancelVoteTxSelector struct {
}

func (CancelStakeTxSelector *CancelVoteTxSelector) Select(tx *types.Transaction) bool {
	return tx.Type() == types.CancelVoteCreditType
}

var (
	_ = (ITransactionSelector)((*CancelVoteTxSelector)(nil))
	_ = (ITransactionValidator)((*CancelVoteTransactionProcessor)(nil))
)

type CancelVoteTransactionProcessor struct {
}

func (processor *CancelVoteTransactionProcessor) ExecuteTransaction(context *ExecuteTransactionContext) *types.ExecuteTransactionResult {
	etr := &types.ExecuteTransactionResult{}
	from := context.From()
	tx := context.Tx()
	stakeStore := context.TrieStore()

	detail, err := stakeStore.CancelVoteCredit(from, tx.To(), tx.Amount(), context.blockContext.Block.Header.Height)
	if err != nil {
		etr.Txerror = err
		return etr
	}

	logs := make([]*types.Log, 0, 1)
	data, _ := json.Marshal(detail)

	log := types.Log{TxType: tx.Type(), TxHash: *tx.TxHash(), Data: data, Height: context.header.Height, TxIndex: 0}
	logs = append(logs, &log)

	err = stakeStore.PutNonce(from, tx.Nonce()+1)
	if err != nil {
		etr.Txerror = err
		return etr
	}
	return etr
}

package chain

import (
	"math/big"

	"github.com/drep-project/DREP-Chain/database"
	types "github.com/drep-project/DREP-Chain/types"
)

type TransactionValidator struct {
	chain *ChainService
}

func NewTransactionValidator(chain *ChainService) *TransactionValidator {
	return &TransactionValidator{
		chain,
	}
}

func (transactionValidator *TransactionValidator) ExecuteTransaction(db *database.Database, tx *types.Transaction, gp *GasPool, header *types.BlockHeader) (*types.Receipt, *big.Int, *big.Int, error) {
	from, err := tx.From()
	if err != nil {
		return nil, nil, nil, err
	}

	gasUsed := new(uint64)
	receipt, _, err := transactionValidator.chain.stateProcessor.ApplyTransaction(db, transactionValidator.chain, gp, header, tx, from, gasUsed)
	if err != nil {
		log.WithField("reason", err).Error("executeTransaction transaction error")
		return nil, nil, nil, err
	}
	gasFee := new(big.Int).Mul(new(big.Int).SetUint64(*gasUsed), tx.GasPrice())
	return receipt, new(big.Int).SetUint64(*gasUsed), gasFee, nil
}

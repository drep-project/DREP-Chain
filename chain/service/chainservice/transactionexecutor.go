package chainservice

import (
	"github.com/drep-project/dlog"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/database"
	"math/big"
	"fmt"
)

type TransactionValidator struct {
	chain *ChainService
}

func NewTransactionValidator(chain *ChainService) *TransactionValidator {
	return &TransactionValidator{
		chain,
	}
}

func (transactionValidator *TransactionValidator) ExecuteTransaction(db *database.Database, tx *chainTypes.Transaction, gp *GasPool, header *chainTypes.BlockHeader) (*chainTypes.Receipt, *big.Int, *big.Int, error) {
	from, err := tx.From()
	fmt.Println("err1: ", err)
	if err != nil {
		return nil, nil, nil, err
	}

	gasUsed := new(uint64)
	receipt, _, err := transactionValidator.chain.stateProcessor.ApplyTransaction(db, transactionValidator.chain, gp, header, tx, from, gasUsed)
	if err != nil {
		dlog.Error("executeTransaction transaction error", "reason", err)
		return nil, nil, nil, err
	}
	gasFee := new(big.Int).Mul(new(big.Int).SetUint64(*gasUsed), tx.GasPrice())
	return receipt, new(big.Int).SetUint64(*gasUsed), gasFee, nil
}

type ITransactionValidator interface {
	ExecuteTransaction(db *database.Database, tx *chainTypes.Transaction, gp *GasPool, header *chainTypes.BlockHeader) (*chainTypes.Receipt, *big.Int, *big.Int, error)
}
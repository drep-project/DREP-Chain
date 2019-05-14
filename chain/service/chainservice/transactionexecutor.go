package service

import (
	"github.com/drep-project/dlog"
	"github.com/drep-project/drep-chain/chain/params"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/pkgs/evm/vm"

	"math"
	"math/big"
	"time"
)

const (
	allowedFutureBlockTime = 15 * time.Second
)

var (
	childTrans []*chainTypes.Transaction
)

func (chainService *ChainService) VerifyTransaction(tx *chainTypes.Transaction) error {
	chainService.stateLock.Lock()
	db := chainService.StateSnapshot.db
	chainService.stateLock.Unlock()
	return chainService.verifyTransaction(db, tx)
}

func (chainService *ChainService) verifyTransaction(db *database.Database, tx *chainTypes.Transaction) error {
	from, err := tx.From()

	// Transactions can't be negative. This may never happen using RLP decoded
	// transactions but may occur if you create a transaction using the RPC.
	if tx.Amount().Sign() < 0 {
		return ErrNegativeAmount
	}

	// Check the transaction doesn't exceed the current
	// block limit gas.
	gasLimit := chainService.BestChain.Tip().GasLimit
	if gasLimit.Uint64() < tx.Gas() {
		return ErrExceedGasLimit
	}

	// Transactor should have enough funds to cover the costs
	// cost == V + GP * GL
	originBalance := db.GetBalance(from)
	if originBalance.Cmp(tx.Cost()) < 0 {
		return ErrBalance
	}

	// Should supply enough intrinsic gas
	gas, err := IntrinsicGas(tx.AsPersistentMessage(), tx.To() == nil || tx.To().IsEmpty())
	if err != nil {
		return err
	}
	if tx.Gas() < gas {
		dlog.Error("gas exceed tx gaslimit ", "gas", gas, "tx.gas", tx.Gas())
		return ErrReachGasLimit
	}
	return nil
}

func (chainService *ChainService) ExecuteTransaction(db *database.Database, tx *chainTypes.Transaction, gp *GasPool, header *chainTypes.BlockHeader) (*big.Int, *big.Int, error) {
	return chainService.executeTransaction(db, tx, gp, header)
}

func (chainService *ChainService) executeTransaction(db *database.Database, tx *chainTypes.Transaction, gp *GasPool, header *chainTypes.BlockHeader) (*big.Int, *big.Int, error) {
	from, err := tx.From()
	if err != nil {
		return nil, nil, err
	}

	gasUsed := new(uint64)
	_, _, err = chainService.stateProcessor.ApplyTransaction(db, chainService, gp, header, tx, from, gasUsed)
	if err != nil {
		dlog.Error("executeTransaction transaction error", "reason", err)
		return nil, nil, err
	}
	gasFee := new(big.Int).Mul(new(big.Int).SetUint64(*gasUsed), tx.GasPrice())
	return new(big.Int).SetUint64(*gasUsed), gasFee, nil
}

func (chainService *ChainService) checkBalance(gaslimit, gasPrice, balance, gasFloor, gasCap *big.Int) error {
	if gasFloor != nil {
		amountFloor := new(big.Int).Mul(gasFloor, gasPrice)
		if gaslimit.Cmp(gasFloor) < 0 || amountFloor.Cmp(balance) > 0 {
			return ErrGas
		}
	}
	if gasCap != nil {
		amountCap := new(big.Int).Mul(gasCap, gasPrice)
		if amountCap.Cmp(balance) > 0 {
			return ErrBalance
		}
	}
	return nil
}

// IntrinsicGas computes the 'intrinsic gas' for a message with the given data.
func IntrinsicGas(data []byte, contractCreation bool) (uint64, error) {
	// Set the starting gas for the raw transaction
	var gas uint64
	if contractCreation {
		gas = params.TxGasContractCreation
	} else {
		gas = params.TxGas
	}
	// Bump the required gas by the amount of transactional data
	if len(data) > 0 {
		// Zero and non-zero bytes are priced differently
		var nz uint64
		for _, byt := range data {
			if byt != 0 {
				nz++
			}
		}
		// Make sure we don't exceed uint64 for all data combinations
		if (math.MaxUint64-gas)/params.TxDataNonZeroGas < nz {
			return 0, vm.ErrOutOfGas
		}
		gas += nz * params.TxDataNonZeroGas

		z := uint64(len(data)) - nz
		if (math.MaxUint64-gas)/params.TxDataZeroGas < z {
			return 0, vm.ErrOutOfGas
		}
		gas += z * params.TxDataZeroGas
	}
	return gas, nil
}
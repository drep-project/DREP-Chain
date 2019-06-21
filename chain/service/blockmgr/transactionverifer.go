package blockmgr

import (
	chainTypes "github.com/drep-project/drep-chain/chain/types"
)

// VerifyTransaction use current tip state as environment may not matched read disk state
// not check tx nonce ; current nonce shoud use pool nonce while receive tx
func (blockMgr *BlockMgr) VerifyTransaction(tx *chainTypes.Transaction) error {
	return blockMgr.verifyTransaction(tx)
}

func (blockMgr *BlockMgr) verifyTransaction(tx *chainTypes.Transaction) error {
	db := blockMgr.ChainService.GetCurrentState()
	from, err := tx.From()

	// Transactions can't be negative. This may never happen using RLP decoded
	// transactions but may occur if you create a transaction using the RPC.
	if tx.Amount().Sign() < 0 {
		return ErrNegativeAmount
	}

	// Check the transaction doesn't exceed the current
	// block limit gas.
	gasLimit := blockMgr.ChainService.BestChain().Tip().GasLimit
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
	gas, err := tx.IntrinsicGas()
	if err != nil {
		return err
	}
	if tx.Gas() < gas {
		log.WithField("gas", gas).WithField("tx.gas", tx.Gas()).Error("gas exceed tx gaslimit ")
		return ErrReachGasLimit
	}
	if tx.Type() == chainTypes.SetAliasType {
		from, err := tx.From()
		if err != nil {
			return err
		}
		alias := blockMgr.ChainService.GetDatabaseService().GetStorageAlias(from)
		if alias != "" {
			return ErrNotSupportRenameAlias
		}
	}
	return nil
}

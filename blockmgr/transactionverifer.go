package blockmgr

import (
	"fmt"
	"github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/types"
)

// VerifyTransaction use current tip state as environment may not matched read disk state
// not check tx nonce ; current nonce shoud use pool nonce while receive tx
func (blockMgr *BlockMgr) verifyTransaction(tx *types.Transaction) error {
	// Transactions can't be negative. This may never happen using RLP decoded
	// transactions but may occur if you create a transaction using the RPC.
	if tx.Amount().Sign() < 0 {
		return ErrNegativeAmount
	}

	tip := blockMgr.ChainService.BestChain().Tip()
	// Check the transaction doesn't exceed the current
	// block limit gas.
	if tip.GasLimit.Uint64() < tx.Gas() {
		return ErrExceedGasLimit
	}

	//Transactor should have enough funds to cover the costs
	trieStore, err := store.TrieStoreFromStore(blockMgr.DatabaseService.LevelDb(), blockMgr.ChainService.BestChain().Tip().StateRoot)
	if err != nil {
		log.WithField("err", err).Trace("verifyTransaction")
		return err
	}

	from, err := tx.From()
	originBalance := trieStore.GetBalance(from, blockMgr.ChainService.BestChain().Height())
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

	return blockMgr.checkByTxType(tx)
}

func (blockMgr *BlockMgr) checkByTxType(tx *types.Transaction) error {
	from, _ := tx.From()
	trieQuery, err := chain.NewTrieQuery(blockMgr.DatabaseService.LevelDb(),
		blockMgr.ChainService.BestChain().Tip().StateRoot)
	if err != nil {
		return err
	}
	switch tx.Type() {
	case types.SetAliasType:
		newAlias := tx.GetData()
		if newAlias == nil {
			return chain.ErrUnsupportAliasChar
		}

		trieStore, err := store.TrieStoreFromStore(blockMgr.DatabaseService.LevelDb(),
			blockMgr.ChainService.BestChain().Tip().StateRoot)
		if err != nil {
			log.WithField("err", err).Trace("check byt tx type")
			return err
		}
		if err := chain.CheckAlias(tx, trieStore, blockMgr.ChainService.BestChain().Height()); err != nil {
			return err
		}

		alias := trieQuery.GetStorageAlias(from)
		if alias != "" {
			return ErrNotSupportRenameAlias
		}
		return nil
	case types.CancelCandidateType, types.CancelVoteCreditType:
		if err = trieQuery.CheckCancelCandidateType(tx); err != nil {
			return err
		}
		return nil
	//case types.CandidateType:
	//	if from.String() == tx.To().String() {
	//		return fmt.Errorf("from euqal to addr")
	//	}
	//	return nil

	case types.VoteCreditType:
		if from.String() == tx.To().String() {
			return fmt.Errorf("from euqal to addr")
		}
		return nil
	case types.TransferType, types.CreateContractType,
		types.CallContractType, types.CandidateType:
		return nil
	}

	return fmt.Errorf("checkByTxType err type:%d", tx.Type())
}

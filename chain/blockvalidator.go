package chain

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"

	"github.com/drep-project/DREP-Chain/chain/block"
	"github.com/drep-project/DREP-Chain/chain/transactions"
	"github.com/drep-project/DREP-Chain/chain/utils"
	"github.com/drep-project/DREP-Chain/crypto"

	"github.com/drep-project/DREP-Chain/common"

	"github.com/drep-project/DREP-Chain/params"
	"github.com/drep-project/DREP-Chain/types"
)

type BlockValidators []IBlockValidator

func (blockValidators BlockValidators) SelectByType(t reflect.Type) IBlockValidator {
	for _, validator := range blockValidators {
		if t == reflect.TypeOf(validator).Elem() {
			return validator
		}
	}
	return nil
}

type IBlockValidator interface {
	VerifyHeader(header, parent *types.BlockHeader) error

	VerifyBody(block *types.Block) error

	ExecuteBlock(context *block.BlockExecuteContext) error
}

//type BlockExecuteContext struct {
//	TrieStore store.StoreInterface
//	Gp        *GasPool
//	DbStore   *ChainStore
//	Block     *types.Block
//	GasUsed   *big.Int
//	GasFee    *big.Int
//	Logs      []*types.Log
//	Receipts  types.Receipts
//}
//
//func NewBlockExecuteContext(trieStore store.StoreInterface, gp *GasPool, dbStore *ChainStore, block *types.Block) *BlockExecuteContext {
//	return &BlockExecuteContext{
//		TrieStore: trieStore,
//		Gp:        gp,
//		DbStore:   dbStore,
//		Block:     block,
//		GasUsed:   new(big.Int),
//		GasFee:    new(big.Int),
//		Logs:      []*types.Log{},
//		Receipts:  types.Receipts{},
//	}
//}
//
//func (blockExecuteContext *BlockExecuteContext) AddGasUsed(gas *big.Int) {
//	blockExecuteContext.GasUsed = blockExecuteContext.GasUsed.Add(blockExecuteContext.GasUsed, gas)
//}
//
//func (blockExecuteContext *BlockExecuteContext) AddGasFee(fee *big.Int) {
//	blockExecuteContext.GasFee = blockExecuteContext.GasFee.Add(blockExecuteContext.GasFee, fee)
//}

type ChainBlockValidator struct {
	chain *ChainService
}

func NewChainBlockValidator(chainService *ChainService) *ChainBlockValidator {
	return &ChainBlockValidator{
		chain: chainService,
	}
}

func (chainBlockValidator *ChainBlockValidator) VerifyHeader(header, parent *types.BlockHeader) error {
	// Verify chainID  matched
	if header.ChainId != chainBlockValidator.chain.ChainID() {
		return ErrChainId
	}
	// Verify version  matched
	if header.Version != common.Version {
		return ErrVersion
	}
	//Verify header's previousHash is equal parent hash
	if header.PreviousHash != *parent.Hash() {
		return ErrPreHash
	}
	// Verify that the block number is parent's +1
	if header.Height-parent.Height != 1 {
		return ErrInvalidateBlockNumber
	}
	// pre block timestamp before this block time
	if header.Timestamp <= parent.Timestamp {
		log.WithField("header height", header.Height).WithField("parent height", parent.Height).WithField("header ts", header.Timestamp).WithField("parent ts", parent.Timestamp).Error("VerifyHeader")
		return ErrInvalidateTimestamp
	}

	// Verify that the gasRemained limit is <= 2^63-1
	cap := uint64(0x7fffffffffffffff)
	if header.GasLimit.Uint64() > cap {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, cap)
	}
	// Verify that the gasRemained is <= gasLimit
	if header.GasUsed.Uint64() > header.GasLimit.Uint64() {
		return fmt.Errorf("invalid gasRemained: have %v, gasLimit %v", header.GasUsed, header.GasLimit)
	}

	//TODO Verify that the gasRemained limit remains within allowed bounds
	nextGasLimit := chainBlockValidator.chain.CalcGasLimit(parent, params.MinGasLimit, params.MaxGasLimit)
	if nextGasLimit.Cmp(&header.GasLimit) != 0 {
		return fmt.Errorf("invalid gasRemained limit: have %v, want %v += %v", header.GasLimit, parent.GasLimit, nextGasLimit)
	}
	return nil
}

func (chainBlockValidator *ChainBlockValidator) VerifyBody(block *types.Block) error {
	// Header validity is known at this point, check the uncles and transactions
	header := block.Header
	if hash := chainBlockValidator.chain.DeriveMerkleRoot(block.Data.TxList); !bytes.Equal(hash, header.TxRoot) {
		return fmt.Errorf("transaction root hash mismatch: have %x, want %x", hash, header.TxRoot)
	}
	return nil
}

func (chainBlockValidator *ChainBlockValidator) ExecuteBlock(context *block.BlockExecuteContext) error {
	context.Receipts = make([]*types.Receipt, context.Block.Data.TxCount)
	context.Logs = make([]*types.Log, 0)
	if len(context.Block.Data.TxList) < 0 {
		return nil
	}

	for i, t := range context.Block.Data.TxList {
		//templateBlockValidator := blockmgr.NewTemplateBlockValidator(chainBlockValidator.chain)
		receipt, gasUsed, err := chainBlockValidator.RouteTransaction(context, context.Gp, t)
		if err != nil {
			return err
		}
		if err == nil {
			gasUsedBig := new(big.Int).SetUint64(gasUsed)
			context.AddGasUsed(gasUsedBig)
			gasFee := new(big.Int).Mul(gasUsedBig, t.GasPrice())
			context.AddGasFee(gasFee)
		} else {
			return err
		}
		context.Receipts[i] = receipt
		context.Logs = append(context.Logs, receipt.Logs...)
	}
	//TODO check whether gasRemained exceed max value
	newReceiptRoot := chainBlockValidator.chain.DeriveReceiptRoot(context.Receipts)
	if newReceiptRoot != context.Block.Header.ReceiptRoot {
		return ErrReceiptRoot
	}

	for _, receipt := range context.Receipts {
		receipt.BlockHash = *context.Block.Header.Hash()
		receipt.PostState = newReceiptRoot[:]
		err := context.DbStore.PutReceipt(receipt.TxHash, receipt)
		if err != nil {
			return err
		}
	}
	err := context.DbStore.PutReceipts(*context.Block.Header.Hash(), context.Receipts)
	if err != nil {
		return err
	}

	return nil
}

func (chainBlockValidator *ChainBlockValidator) RouteTransaction(context *block.BlockExecuteContext, gasPool *utils.GasPool, tx *types.Transaction) (*types.Receipt, uint64, error) {
	//init transaction tx
	from, err := tx.From()
	if err != nil {
		return nil, 0, err
	}

	txContext := transactions.NewExecuteTransactionContext(context, context.TrieStore, gasPool, from, tx)
	if err := txContext.PreCheck(); err != nil {
		return nil, 0, err
	}

	// Pay intrinsic gastx
	gas, err := tx.IntrinsicGas()
	if err != nil {
		return nil, 0, err
	}

	if err = txContext.UseGas(gas); err != nil {
		return nil, 0, err
	}

	process := &transactions.Processor{}
	ret := process.ExecuteTransaction(txContext)
	if ret.Txerror != nil {
		return nil, 0, ret.Txerror
	}
	err = txContext.RefundCoin()
	if err != nil {
		return nil, 0, err
	}
	// Create a new receipt for the transaction, storing the intermediate root and gasRemained used by the tx
	// based on the eip phase, we're passing whether the root touch-delete accounts.
	receipt := types.NewReceipt(crypto.ZeroHash[:], ret.ContractTxExecuteFail, txContext.GasUsed())
	receipt.TxHash = *tx.TxHash()
	receipt.GasUsed = txContext.GasUsed()
	// if the transaction created a contract, store the creation address in the receipt.
	if (tx.To() == nil || tx.To().IsEmpty()) && tx.Type() == types.CreateContractType {
		receipt.ContractAddress = crypto.CreateAddress(*from, tx.Nonce())
		log.WithField("contractAddr:", receipt.ContractAddress).Info("execute contract")
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = ret.ContractTxLog
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
	//receipt.BlockHash = *header.Hash()
	receipt.BlockNumber = context.Block.Header.Height
	return receipt, txContext.GasUsed(), nil

	panic("never come here")
}

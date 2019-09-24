package chain

import (
	"bytes"
	"fmt"
	"github.com/drep-project/drep-chain/crypto"
	"math/big"

	"github.com/drep-project/drep-chain/common"

	"github.com/drep-project/drep-chain/params"
	"github.com/drep-project/drep-chain/types"
)

type IBlockValidator interface {
	VerifyHeader(header, parent *types.BlockHeader) error

	VerifyBody(block *types.Block) error

	ExecuteBlock(context *BlockExecuteContext) error
}

type BlockExecuteContext struct {
	TrieStore *TrieStore
	Gp        *GasPool
	DbStore   *ChainStore
	Block     *types.Block
	GasUsed   *big.Int
	GasFee    *big.Int
	Logs      []*types.Log
	Receipts  types.Receipts
}

func NewBlockExecuteContext(trieStore *TrieStore, gp *GasPool, dbStore *ChainStore, block *types.Block) *BlockExecuteContext {
	return &BlockExecuteContext{
		TrieStore: trieStore,
		Gp: gp,
		DbStore: dbStore,
		Block: block,
		GasUsed: new(big.Int),
		GasFee:new(big.Int),
		Logs:[]*types.Log{},
		Receipts: types.Receipts{},
	}
}

func (blockExecuteContext *BlockExecuteContext) AddGasUsed(gas *big.Int) {
	blockExecuteContext.GasUsed = blockExecuteContext.GasUsed.Add(blockExecuteContext.GasUsed, gas)
}

func (blockExecuteContext *BlockExecuteContext) AddGasFee(fee *big.Int) {
	blockExecuteContext.GasFee = blockExecuteContext.GasFee.Add(blockExecuteContext.GasFee, fee)
}

type ChainBlockValidator struct {
	chain *ChainService
}

func NewChainBlockValidator(chainService *ChainService) *ChainBlockValidator {
	return &ChainBlockValidator{
		chain: chainService,
	}
}

func (chainBlockValidator *ChainBlockValidator) VerifyHeader(header, parent *types.BlockHeader) error {
	// Verify chainId  matched
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

func (chainBlockValidator *ChainBlockValidator) ExecuteBlock(context *BlockExecuteContext) error {
	context.Receipts = make([]*types.Receipt, context.Block.Data.TxCount)
	context.Logs = make([]*types.Log, 0)
	if len(context.Block.Data.TxList) < 0 {
		return nil
	}

	for i, t := range context.Block.Data.TxList {
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
	err := context.DbStore.PutReceipts(*context.Block.Header.Hash(), context.Receipts)
	if err != nil {
		return err
	}
	for _, receipt := range context.Receipts {
		receipt.PostState = newReceiptRoot[:]
		err =context.DbStore.PutReceipt(receipt.TxHash, receipt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (chainBlockValidator *ChainBlockValidator) RouteTransaction(context *BlockExecuteContext, gasPool *GasPool, tx *types.Transaction) (*types.Receipt, uint64, error) {
	//init transaction tx
	from, err := tx.From()
	if err != nil {
		return nil, 0, err
	}

	txContext := NewExecuteTransactionContext(context, context.TrieStore, gasPool, from, tx)
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

	exit := false
	for selector, txValidator := range chainBlockValidator.chain.transactionValidator {
		if selector.Select(tx) {
			exit = true
			_, failed, logs, err := txValidator.ExecuteTransaction(txContext)
			if err != nil {
				return nil, 0, err
			}
			err = txContext.RefundCoin()
			if err != nil {
				return nil, 0, err
			}
			//context.TrieStore.CacheToTrie()
			// Create a new receipt for the transaction, storing the intermediate root and gasRemained used by the tx
			// based on the eip phase, we're passing whether the root touch-delete accounts.
			//crypto.ZeroHash[:]
			receipt := types.NewReceipt(crypto.ZeroHash[:], failed, txContext.GasUsed())
			receipt.TxHash = *tx.TxHash()
			receipt.GasUsed = txContext.GasUsed()
			// if the transaction created a contract, store the creation address in the receipt.
			if tx.To() == nil || tx.To().IsEmpty() {
				receipt.ContractAddress = crypto.CreateAddress(*from, tx.Nonce())
				fmt.Println(receipt.ContractAddress)
			}
			// Set the receipt logs and create a bloom for filtering
			receipt.Logs = logs
			receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
			//receipt.BlockHash = *header.Hash()
			receipt.BlockNumber = context.Block.Header.Height
			return receipt, txContext.GasUsed(), nil
		}
	}
	if !exit {
		return nil, 0, ErrUnsupportTxType
	}
	panic("never come here")
}

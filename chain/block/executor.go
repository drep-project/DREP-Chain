package block

import (
	"math/big"

	"github.com/drep-project/DREP-Chain/chain/utils"

	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/types"
)

type BlockExecuteContext struct {
	TrieStore store.StoreInterface
	Gp        *utils.GasPool
	DbStore   *store.ChainStore
	Block     *types.Block
	GasUsed   *big.Int
	GasFee    *big.Int
	Logs      []*types.Log
	Receipts  types.Receipts
}

func NewBlockExecuteContext(trieStore store.StoreInterface, gp *utils.GasPool, dbStore *store.ChainStore, block *types.Block) *BlockExecuteContext {
	return &BlockExecuteContext{
		TrieStore: trieStore,
		Gp:        gp,
		DbStore:   dbStore,
		Block:     block,
		GasUsed:   new(big.Int),
		GasFee:    new(big.Int),
		Logs:      []*types.Log{},
		Receipts:  types.Receipts{},
	}
}

func (blockExecuteContext *BlockExecuteContext) AddGasUsed(gas *big.Int) {
	blockExecuteContext.GasUsed = blockExecuteContext.GasUsed.Add(blockExecuteContext.GasUsed, gas)
}

func (blockExecuteContext *BlockExecuteContext) AddGasFee(fee *big.Int) {
	blockExecuteContext.GasFee = blockExecuteContext.GasFee.Add(blockExecuteContext.GasFee, fee)
}

package blockmgr

import (
	"math/big"
	"time"

	"github.com/drep-project/DREP-Chain/chain/utils"

	chainBlock "github.com/drep-project/DREP-Chain/chain/block"
	chainStore "github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/params"
	"github.com/drep-project/DREP-Chain/types"
)

// GenerateTemplate blockchain t
func (blockMgr *BlockMgr) GenerateTemplate(trieStore chainStore.StoreInterface, leaderAddr crypto.CommonAddress, blockInterval int) (*types.Block, *big.Int, error) {
	parent, err := blockMgr.ChainService.GetHighestBlock()
	if err != nil {
		return nil, nil, err
	}
	newGasLimit := blockMgr.ChainService.CalcGasLimit(parent.Header, params.MinGasLimit, params.MaxGasLimit)
	height := blockMgr.ChainService.BestChain().Height() + 1
	txs := blockMgr.transactionPool.GetPending(newGasLimit)
	previousHash := blockMgr.ChainService.BestChain().Tip().Hash
	timestamp := uint64(time.Now().Unix())

	blockHeader := &types.BlockHeader{
		Version:      common.Version,
		PreviousHash: *previousHash,
		ChainId:      blockMgr.ChainService.ChainID(),
		GasLimit:     *newGasLimit,
		Timestamp:    timestamp,
		Height:       height,
		StateRoot:    []byte{},
		TxRoot:       []byte{},
		MinerAddr:    leaderAddr,
	}

	block := &types.Block{
		Header: blockHeader,
		Data: &types.BlockData{
			TxCount: uint64(len(txs)),
			TxList:  txs,
		},
	}

	gp := new(utils.GasPool).AddGas(newGasLimit.Uint64())
	//process transaction
	chainStore := &chainStore.ChainStore{blockMgr.DatabaseService.LevelDb()}
	context := chainBlock.NewBlockExecuteContext(trieStore, gp, chainStore, block)

	templateValidator := NewTemplateBlockValidator(blockMgr.ChainService)
	err = templateValidator.ExecuteBlock(context, blockInterval)
	if err != nil {
		return nil, nil, err
	}
	return context.Block, context.GasFee, nil
}

package blockmgr

import (
	"fmt"
	"math/big"
	"time"

	"github.com/drep-project/drep-chain/crypto"

	"github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/params"
	"github.com/drep-project/drep-chain/types"
)

func (blockMgr *BlockMgr) GenerateBlock(db *database.Database, leaderAddr crypto.CommonAddress) (*types.Block, *big.Int, error) {
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
		Version:       common.Version,
		PreviousHash:  *previousHash,
		ChainId:       blockMgr.ChainService.ChainID(),
		GasLimit:      *newGasLimit,
		Timestamp:     timestamp,
		Height:        height,
		StateRoot:     []byte{},
		TxRoot:        []byte{},
	}

	finalTxs := make([]*types.Transaction, 0, len(txs))
	finalReceipts := make([]*types.Receipt, 0, len(txs))

	gasUsed := new(big.Int)
	gasFee := new(big.Int)
	gp := new(chain.GasPool).AddGas(blockHeader.GasLimit.Uint64())
	stopchanel := make(chan struct{}, 1)
	tm := time.AfterFunc(time.Second*5, func() {
		stopchanel <- struct{}{}
	})

SELECT_TX:
	for _, t := range txs {
		snap := db.CopyState()
		fmt.Println(gp)
		newGp := *gp
		select {
		case <-stopchanel:
			break SELECT_TX
		default:
			receipt, txGasUsed, txGasFee, err := blockMgr.ChainService.TransactionValidator().ExecuteTransaction(db, t, &newGp, blockHeader)
			if err == nil {
				finalTxs = append(finalTxs, t)
				finalReceipts = append(finalReceipts, receipt)
				gasUsed.Add(gasUsed, txGasUsed)
				gasFee.Add(gasFee, txGasFee)
				gp = &newGp // use new gp and new state if success
			} else {
				//revert old state and use old gp if fail
				db.RevertState(snap)
				if err.Error() == ErrReachGasLimit.Error() {
					break SELECT_TX
				} else {
					log.WithField("Reason", err).Warn("generate block fail")
					continue
				}
			}
		}
	}
	tm.Stop()
	blockHeader.GasUsed = *new(big.Int).SetUint64(gasUsed.Uint64())
	blockHeader.TxRoot = blockMgr.ChainService.DeriveMerkleRoot(finalTxs)
	blockHeader.ReceiptRoot = blockMgr.ChainService.DeriveReceiptRoot(finalReceipts)

	if len(finalReceipts) != 0 {
		blockHeader.Bloom = types.CreateBloom(finalReceipts)
	}

	block := &types.Block{
		Header: blockHeader,
		Data: &types.BlockData{
			TxCount: uint64(len(finalTxs)),
			TxList:  finalTxs,
		},
	}
	return block, gasFee, nil
}

package blockmgr

import (
	"fmt"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/database"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/chain/service/chainservice"
	 "github.com/drep-project/drep-chain/chain/params"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/dlog"
	"time"
	"math/big"
)

func (blockMgr *BlockMgr) GenerateBlock(db *database.Database,leaderKey *secp256k1.PublicKey) (*chainTypes.Block, *big.Int,error) {
	parent, err := blockMgr.ChainService.GetHighestBlock()
	if err != nil {
		return nil, nil, err
	}
	newGasLimit := blockMgr.ChainService.CalcGasLimit(parent.Header, params.MinGasLimit, params.MaxGasLimit)
	height := blockMgr.ChainService.BestChain.Height() + 1
	txs := blockMgr.transactionPool.GetPending(newGasLimit)
	previousHash := blockMgr.ChainService.BestChain.Tip().Hash
	fmt.Println(previousHash)
	fmt.Println(parent.Header.Hash())
	timestamp := uint64(time.Now().Unix())

	blockHeader := &chainTypes.BlockHeader{
		Version:      	common.Version,
		PreviousHash: 	*previousHash,
		ChainId:      	blockMgr.ChainService.ChainID(),
		GasLimit:     	*newGasLimit,
		Timestamp:    	timestamp,
		Height:       	height,
		StateRoot:  	[]byte{},
		TxRoot: 		[]byte{},
		LeaderPubKey: 	*leaderKey,
	}

	finalTxs := make([]*chainTypes.Transaction, 0, len(txs))
	gasUsed := new(big.Int)
	gasFee := new (big.Int)
	gp := new(chainservice.GasPool).AddGas(blockHeader.GasLimit.Uint64())
	stopchanel := make(chan struct{})
	time.AfterFunc(time.Second*5, func() {
		stopchanel <- struct{}{}
	})

SELECT_TX:
	for _, t := range txs {
		select {
		case <-stopchanel:
			break SELECT_TX
		default:
			txGasUsed, txGasFee, err := blockMgr.ChainService.TransactionValidator.ExecuteTransaction(db, t, gp, blockHeader)
			if err == nil {
				finalTxs = append(finalTxs, t)
				gasUsed.Add(gasUsed, txGasUsed)
				gasFee.Add(gasFee, txGasFee)
			} else {
				if err.Error() == ErrReachGasLimit.Error() {
					break SELECT_TX
				} else {
					dlog.Warn("generate block", "exe tx err", err)
					continue
					//  return nil, err
				}
			}
		}
	}

	blockHeader.GasUsed = *new(big.Int).SetUint64(gasUsed.Uint64())
	blockHeader.TxRoot = blockMgr.ChainService.DeriveMerkleRoot(finalTxs)

	block := &chainTypes.Block{
		Header: blockHeader,
		Data: &chainTypes.BlockData{
			TxCount: uint64(len(finalTxs)),
			TxList:  finalTxs,
		},
	}
	return block, gasFee, nil
}
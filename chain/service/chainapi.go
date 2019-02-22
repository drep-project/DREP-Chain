package service

import (
    "math/big"
    chainType "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/drep-chain/database"
    "github.com/drep-project/drep-chain/crypto"
    "github.com/drep-project/drep-chain/common"
)

var dbService = &database.DatabaseService{}

type ChainApi struct {
    chain *ChainService
}

func (chain *ChainApi) GetBlock(height int64) *chainType.Block {
    if height < 0 {
        return nil
    }
    return dbService.GetBlock(height)
}

func (chain *ChainApi) GetMaxHeight() int64 {
    return dbService.GetMaxHeight()
}

func (chain *ChainApi) GetBalance(addr crypto.CommonAddress, chainId common.ChainIdType) *big.Int{
    return dbService.GetBalance(addr, chainId, true)
}

func (chain *ChainApi) GetNonce(addr crypto.CommonAddress, chainId common.ChainIdType) int64 {
    return dbService.GetNonce(addr, chainId, true)
}

func (chain *ChainApi) GetPreviousBlockHash() string {
    bytes := dbService.GetPreviousBlockHash()
    return "0x" + string(bytes)
}

func (chain *ChainApi) GetReputation(addr crypto.CommonAddress, chainId common.ChainIdType) string {
    rep := dbService.GetReputation(addr, chainId, true)
    return rep.String()
}

func (chain *ChainApi) GetTransactionsFromBlock(height int64) []*chainType.Transaction {
    block := dbService.GetBlock(height)
    return block.Data.TxList
}

func (chain *ChainApi) GetTransactionByBlockHeightAndIndex(height int64, index int) *chainType.Transaction{
    block := dbService.GetBlock(height)
    if index > len(block.Data.TxList) {
        return nil
    }
    return block.Data.TxList[index]
}

func (chain *ChainApi) GetTransactionCountByBlockHeight(height int64) int {
    block := dbService.GetBlock(height)
    return len(block.Data.TxList)
}
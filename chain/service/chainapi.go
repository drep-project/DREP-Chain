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

func (chain *ChainApi) GetBalance(addr string, chainId string) *big.Int{
    if addr == "" {
        return big.NewInt(0)
    }
    address := crypto.String2Address(addr)
    chainid := common.String2ChainId(chainId)

    return dbService.GetBalance(address, chainid, true)
}

func (chain *ChainApi) GetNonce(addr string, chainId string) int64 {
    if addr == "" {
        return 0
    }
    address := crypto.String2Address(addr)
    chainid := common.String2ChainId(chainId)

    return dbService.GetNonce(address, chainid, true)
}

func (chain *ChainApi) GetPreviousBlockHash() string {
    bytes := dbService.GetPreviousBlockHash()
    return "0x" + string(bytes)
}

func (chain *ChainApi) GetReputation(addr string, chainId string) string {
    if addr == "" {
        return ""
    }
    address := crypto.String2Address(addr)
    chainid := common.String2ChainId(chainId)
    rep := dbService.GetReputation(address, chainid, true)
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

//TODO mock a rpc to provent rpc error
func (chainApi *ChainApi) Mock(){

}
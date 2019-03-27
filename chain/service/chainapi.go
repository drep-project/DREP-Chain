package service

import (
    chainType "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/drep-chain/database"
    "math/big"
)


type ChainApi struct {
    chainService *ChainService
    dbService *database.DatabaseService `service:"database"`
}

func (chain *ChainApi) GetBlock(height int64) *chainType.Block {
    blocks, _ := chain.chainService.GetBlocksFrom(height, 1)
    return blocks[0]
}

func (chain *ChainApi) GetMaxHeight() int64 {
    return chain.chainService.BestChain.Height()
}

func (chain *ChainApi) GetBalance(accountName string) *big.Int{
    return chain.dbService.GetBalance(accountName, true)
}

func (chain *ChainApi) GetAccount(accountName string) (*chainType.Storage, error) {
    return chain.dbService.GetStorage(accountName, true)
}

func (chain *ChainApi) GetNonce(accountName string) int64 {
    return chain.dbService.GetNonce(accountName, true)
}

func (chain *ChainApi) GetReputation(accountName string) *big.Int {
    return chain.dbService.GetReputation(accountName, true)
}

func (chain *ChainApi) GetTransactionsFromBlock(height int64) []*chainType.Transaction {
    block := chain.GetBlock(height)
    return block.Data.TxList
}

func (chain *ChainApi) GetTransactionByBlockHeightAndIndex(height int64, index int) *chainType.Transaction{
    block := chain.GetBlock(height)
    if index > len(block.Data.TxList) {
        return nil
    }
    return block.Data.TxList[index]
}

func (chain *ChainApi) GetTransactionCountByBlockHeight(height int64) int {
    block := chain.GetBlock(height)
    return len(block.Data.TxList)
}

func (chain *ChainApi) SendRawTransaction(tx *chainType.Transaction) (string, error){
    err := chain.chainService.ValidateTransaction(tx)
    if err != nil {
        return "", err
    }
    err = chain.chainService.transactionPool.AddTransaction(tx)
    if err != nil {
        return "", err
    }

    chain.chainService.P2pServer.Broadcast(tx)
    return tx.TxHash().String(), err
}

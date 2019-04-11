package service

import (
    chainType "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/drep-chain/crypto"
    "github.com/drep-project/drep-chain/database"
    "math/big"
)


type ChainApi struct {
    chainService *ChainService
    dbService *database.DatabaseService `service:"database"`
}

func (chain *ChainApi) GetBlock(height uint64) *chainType.Block {
    blocks, _ := chain.chainService.GetBlocksFrom(height, 1)
    return blocks[0]
}

func (chain *ChainApi) GetMaxHeight() uint64 {
    return chain.chainService.BestChain.Height()
}

func (chain *ChainApi) GetBalance(addr crypto.CommonAddress) *big.Int{
    return chain.dbService.GetBalance(&addr, true)
}

func (chain *ChainApi) GetNonce(addr crypto.CommonAddress) uint64 {
    return chain.dbService.GetNonce(&addr, true)
}

func (chain *ChainApi) GetPreviousBlockHash() string {
    block := chain.GetBlock(chain.GetMaxHeight())
    hash := block.Header.PreviousHash.String()
    return "0x" + hash
}

func (chain *ChainApi) GetReputation(addr crypto.CommonAddress) *big.Int {
    return chain.dbService.GetReputation(&addr, true)
}

func (chain *ChainApi) GetTransactionsFromBlock(height uint64) []*chainType.Transaction {
    block := chain.GetBlock(height)
    return block.Data.TxList
}

func (chain *ChainApi) GetTransactionByBlockHeightAndIndex(height uint64, index int) *chainType.Transaction{
    block := chain.GetBlock(height)
    if index > len(block.Data.TxList) {
        return nil
    }
    return block.Data.TxList[index]
}

func (chain *ChainApi) GetTransactionCountByBlockHeight(height uint64) int {
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

    chain.chainService.BroadcastTx(chainType.MsgTypeTransaction, tx, true)

    return tx.TxHash().String(), err
}
package service

import (
    "math/big"
    chainType "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/drep-chain/database"
    "github.com/drep-project/drep-chain/crypto"
    "github.com/drep-project/drep-chain/common"
    "fmt"
)


type ChainApi struct {
    chainService *ChainService
    dbService *database.DatabaseService `service:"database"`
}

func (chain *ChainApi) GetBlock(height int64) *chainType.Block {
    if height < 0 {
        return nil
    }
    return chain.dbService.GetBlock(height)
}

func (chain *ChainApi) GetMaxHeight() int64 {
    return chain.dbService.GetMaxHeight()
}

func (chain *ChainApi) GetBalance(addr crypto.CommonAddress, chainId common.ChainIdType) *big.Int{
    bbb := chain.dbService.GetBalance(addr, common.ChainIdType{}, true)
    text, _ := chainId.MarshalText()
    fmt.Println(string(text))

    text, _ = addr.MarshalText()
    fmt.Println(string(text))
    fmt.Println(common.ChainIdType{}.Hex())

    fmt.Println(bbb.Int64())
    return bbb
}

func (chain *ChainApi) GetNonce(addr crypto.CommonAddress, chainId common.ChainIdType) int64 {
    return chain.dbService.GetNonce(addr, true)
}

func (chain *ChainApi) GetPreviousBlockHash() string {
    bytes := chain.dbService.GetPreviousBlockHash()
    return "0x" + string(bytes)
}

func (chain *ChainApi) GetReputation(addr crypto.CommonAddress, chainId common.ChainIdType) *big.Int {
    return chain.dbService.GetReputation(addr, chainId, true)
}

func (chain *ChainApi) GetTransactionsFromBlock(height int64) []*chainType.Transaction {
    block := chain.dbService.GetBlock(height)
    return block.Data.TxList
}

func (chain *ChainApi) GetTransactionByBlockHeightAndIndex(height int64, index int) *chainType.Transaction{
    block := chain.dbService.GetBlock(height)
    if index > len(block.Data.TxList) {
        return nil
    }
    return block.Data.TxList[index]
}

func (chain *ChainApi) GetTransactionCountByBlockHeight(height int64) int {
    block := chain.dbService.GetBlock(height)
    return len(block.Data.TxList)
}

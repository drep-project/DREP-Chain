package service

import (
    chainType "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/drep-chain/common"
    "github.com/drep-project/drep-chain/crypto"
    "github.com/drep-project/binary"
    "github.com/drep-project/drep-chain/database"
    "math/big"
)


type ChainApi struct {
    chainService *ChainService
    dbService *database.DatabaseService `service:"database"`
}

func (chain *ChainApi) GetBlock(height uint64) *chainType.RpcBlock  {
    blocks, _ := chain.chainService.GetBlocksFrom(height, 1)
    return  new (chainType.RpcBlock).From(blocks[0])
}

func (chain *ChainApi) GetMaxHeight() uint64 {
    return chain.chainService.BestChain.Height()
}

func (chain *ChainApi) GetBalance(addr crypto.CommonAddress) *big.Int{
    return chain.dbService.GetBalance(&addr, false)
}

func (chain *ChainApi) GetNonce(addr crypto.CommonAddress) uint64 {
    return chain.dbService.GetNonce(&addr, false)
}

func (chain *ChainApi) GetPreviousBlockHash() string {
    block := chain.GetBlock(chain.GetMaxHeight())
    return block.PreviousHash.String()
}

func (chain *ChainApi) GetReputation(addr crypto.CommonAddress) *big.Int {
    return chain.dbService.GetReputation(&addr, false)
}

func (chain *ChainApi) GetTransactionsFromBlock(height uint64) []*chainType.RpcTransaction  {
    block := chain.GetBlock(height)
    return block.Txs
}

func (chain *ChainApi) GetTransactionByBlockHeightAndIndex(height uint64, index int) *chainType.RpcTransaction {
    block := chain.GetBlock(height)
    if index > len(block.Txs) {
        return nil
    }
    return block.Txs[index]
}

func (chain *ChainApi) GetTransactionCountByBlockHeight(height uint64) int {
    block := chain.GetBlock(height)
    return len(block.Txs)
}

func (chain *ChainApi) SendRawTransaction(txbytes common.Bytes) (string, error){
    tx := &chainType.Transaction{}
    err := binary.Unmarshal(txbytes,tx)
    if err != nil {
        return "", err
    }

    err = chain.chainService.ValidateTransaction(tx)
    if err != nil {
        return "", err
    }

    chain.chainService.BroadcastTx(chainType.MsgTypeTransaction, tx, true)

    return tx.TxHash().String(), err
}
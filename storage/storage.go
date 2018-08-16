package storage

import (
    "BlockChainTest/common"
    "math/big"
    "sync"
    "BlockChainTest/pool"
)

var (
    balances map[common.Address]big.Int
    balancesLock sync.Locker
    currentHeight int32
)
func getBalance(addr common.Address) big.Int {
    return balances[addr]
}

func SetBalance(addr common.Address, bal big.Int) {
    balancesLock.Lock()
    balances[addr] = bal
    balancesLock.Unlock()
}

func proceed() {
    blocks := pool.GetBlocks(currentHeight)
    for _, v := range blocks {
        executeTransactions(v)
    }
}

func executeTransactions(b *common.Block) {
    for _, t := range b.Data.TxList {
        execute(t)
    }
}

func execute(t *common.Transaction)  {
}
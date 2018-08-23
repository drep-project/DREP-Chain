package store

import (
    "BlockChainTest/common"
    "math/big"
    "sync"
)

var (
    balances map[common.Address]big.Int
    balancesLock sync.Locker
)

func getBalance(addr common.Address) big.Int {
    return balances[addr]
}

func SetBalance(addr common.Address, bal big.Int) {
    balancesLock.Lock()
    balances[addr] = bal
    balancesLock.Unlock()
}

func executeTransactions(b *common.Block) {
    for _, t := range b.Data.TxList {
        execute(t)
    }
}

func execute(t *common.Transaction)  {
}
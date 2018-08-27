package store

import (
    "math/big"
    "sync"
    "BlockChainTest/bean"
)

var (
    balances map[bean.Address]big.Int
    balancesLock sync.Locker
)

func getBalance(addr bean.Address) big.Int {
    return balances[addr]
}

func SetBalance(addr bean.Address, bal big.Int) {
    balancesLock.Lock()
    balances[addr] = bal
    balancesLock.Unlock()
}

func executeTransactions(b *bean.Block) {
    for _, t := range b.Data.TxList {
        execute(t)
    }
}

func execute(t *bean.Transaction)  {
}
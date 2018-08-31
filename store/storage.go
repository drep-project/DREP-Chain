package store

import (
    "math/big"
    "sync"
    "BlockChainTest/bean"
)

var (
    balances map[bean.Address]big.Int
    balancesLock sync.Locker
    currentBlockHeight int64 = 0
)

func getBalance(addr bean.Address) big.Int {
    return balances[addr]
}

func SetBalance(addr bean.Address, bal big.Int) {
    balancesLock.Lock()
    balances[addr] = bal
    balancesLock.Unlock()
}

func ExecuteTransactions(b *bean.Block) {
    if b == nil || b.Header == nil { // || b.Data == nil || b.Data.TxList == nil {
        return
    }
    currentBlockHeight = b.Header.Height
    if b.Data == nil || b.Data.TxList == nil {
        return
    }
    for _, t := range b.Data.TxList {
        execute(t)
    }
}

func execute(t *bean.Transaction)  {
}

func GetCurrentBlockHeight() int64 {
    return currentBlockHeight
}
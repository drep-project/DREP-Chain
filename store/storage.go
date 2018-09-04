package store

import (
    "math/big"
    "sync"
    "BlockChainTest/bean"
)

var (
    balances           map[bean.Address]*big.Int
    nonces             map[bean.Address]int64
    accountLock        sync.Mutex
    currentBlockHeight int64 = 0
    one = big.NewInt(1)
)

func GetBalance(addr bean.Address) *big.Int {
    return balances[addr]
}

func SetBalance(addr bean.Address, bal *big.Int) {
    accountLock.Lock()
    balances[addr] = bal
    accountLock.Unlock()
}

func GetNonce(addr bean.Address) int64 {
    return nonces[addr]
}

func AddNonce(addr bean.Address) {
    accountLock.Lock()
    value, exists := nonces[addr]
    if exists {
        if value >= 0 {
            nonces[addr]++
        } else {
            nonces[addr] = 1
        }
    } else {
        nonces[addr] = 1
    }
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
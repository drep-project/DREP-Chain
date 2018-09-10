package store

import (
    "math/big"
    "sync"
    "BlockChainTest/bean"
    "log"
)

var (
    balances           map[bean.Address]*big.Int
    nonces             map[bean.Address]int64
    accountLock        sync.Mutex
    currentBlockHeight int64 = 0
    one = big.NewInt(1)
)

func GetBalance(addr bean.Address) *big.Int {
    balance, exists := balances[addr]
    if exists {
        return balance
    } else {
        balance = big.NewInt(0)
        balances[addr] = balance
        return balance
    }
}

func GetNonce(addr bean.Address) int64 {
    return nonces[addr]
}

func addNonce(addr bean.Address) {
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

func execute(t *bean.Transaction) *big.Int {
    addr := t.Addr()
    nonce := t.Data.Nonce
    curN := GetNonce(addr)
    if curN + 1 != nonce {
        return nil
    }
    addNonce(addr)
    gasPrice := big.NewInt(0).SetBytes(t.Data.GasPrice)
    gasLimit := big.NewInt(0).SetBytes(t.Data.GasLimit)
    gasFee := big.NewInt(0).Mul(gasLimit, gasPrice)
    balance := GetBalance(addr)
    if gasFee.Cmp(balance) > 0 {
        log.Fatal("Error, gas not right")
        return nil
    }
    if gasLimit.Cmp(TransferGas) < 0 {
        balance.Sub(balance, gasFee)
    } else {
        amount := big.NewInt(0).SetBytes(t.Data.Amount)
        total := big.NewInt(0).Add(amount, TransferGas)
        if balance.Cmp(total) >= 0 {
            balance.Sub(balance, total)
            to := bean.Address(t.Data.To)
            balance2 := GetBalance(to)
            balance2.Add(balance2, amount)
        } else {
            balance.Sub(balance, gasFee)
        }
    }
    return gasFee
}

func GetCurrentBlockHeight() int64 {
    return currentBlockHeight
}
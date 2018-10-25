package store

import (
    "math/big"
    "sync"
    "BlockChainTest/bean"
    "log"
    "fmt"
    "BlockChainTest/database"
)

var (
    balances           = make(map[bean.Address]*big.Int)//map[bean.Address]*big.Int
    nonces             map[bean.Address]int64
    //blocks             []*bean.Block
    accountLock        sync.Mutex
    //currentBlockHeight int64 = -1
)

func init()  {
    //balances = make(map[bean.Address]*big.Int)
    nonces = make(map[bean.Address]int64)
    //blocks = make([]*bean.Block, 0)
}

func GetBalance(addr bean.Address) *big.Int {
    accountLock.Lock()
    defer accountLock.Unlock()
    balance, exists := balances[addr]
    if exists {
        // TODO if map is nil what the fuck
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
    accountLock.Unlock()
}

func ExecuteTransactions(b *bean.Block, del bool) *big.Int {
    if b == nil || b.Header == nil { // || b.Data == nil || b.Data.TxList == nil {
        fmt.Errorf("error block nil or header nil")
        return nil
    }
    height := GetCurrentBlockHeight()
    if height + 1 != b.Header.Height {
        fmt.Println("error", height, b.Header.Height)
        return nil
    }
    // TODO check height
    height = b.Header.Height
    database.PutInt("height", int(height))
    //blocks = append(blocks, b)
    database.SaveBlock(b)
    total := big.NewInt(0)
    if b.Data == nil || b.Data.TxList == nil {
        return total
    }
    for _, t := range b.Data.TxList {
        gasFee := execute(t)
        if del {
            fmt.Println("Delete transaction ", *t)
            removeTransaction(t)
        } else {
            fmt.Println("Does not delete transaction ", *t)
        }
        if gasFee != nil {
            total.Add(total, gasFee)
        }
    }
    return total
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
    switch t.Data.Type {
    case TransferType:
        {
            if gasLimit.Cmp(TransferGas) < 0 {
                balance.Sub(balance, gasFee)
            } else {
                amount := big.NewInt(0).SetBytes(t.Data.Amount)
                total := big.NewInt(0).Add(amount, gasFee)
                if balance.Cmp(total) >= 0 {
                    balance.Sub(balance, total)
                    to := bean.Address(t.Data.To)
                    balance2 := GetBalance(to)
                    balance2.Add(balance2, amount)
                } else {
                    balance.Sub(balance, gasFee)
                }
            }
        }
    case MinerType:
        {
            // TODO if not the admin
            if gasLimit.Cmp(MinerGas) < 0 {
                balance.Sub(balance, gasFee)
            } else {
                balance.Sub(balance, gasFee)
                AddMiner(bean.Address(t.Data.Data))
            }
        }
    }
    return gasFee
}

func GetCurrentBlockHeight() int64 {
    if height, err := database.GetInt("Height"); err == nil {
        return int64(height)
    } else {
        return -1
    }
}

func GetBlocks(from int64, number int64) []*bean.Block {
    bs := database.LoadAllBlock(0)
    l := int64(len(bs))
    if l - 1 < from {
        return []*bean.Block{}
    }
    end := from + number - 1
    r := make([]*bean.Block, 0)
    for i := from; i < l && i <= end; i++ {
        r = append(r, bs[i])
    }
    return r
}
package store

import (
    "math/big"
    "BlockChainTest/bean"
    "log"
    "fmt"
    "BlockChainTest/database"
    "BlockChainTest/accounts"
)

func ExecuteTransactions(b *bean.Block) *big.Int {
    if b == nil || b.Header == nil { // || b.Data == nil || b.Data.TxList == nil {
        fmt.Errorf("error block nil or header nil")
        return nil
    }
    height := database.GetMaxHeight()
    if height + 1 != b.Header.Height {
        fmt.Println("error", height, b.Header.Height)
        return nil
    }
    // TODO check height
    height = b.Header.Height
    database.PutMaxHeight(height)
    //blocks = append(blocks, b)
    database.PutBlock(b)
    total := big.NewInt(0)
    if b.Data == nil || b.Data.TxList == nil {
        return total
    }
    for _, t := range b.Data.TxList {
        gasFee := execute(t)
        fmt.Println("Delete transaction ", *t)
        fmt.Println(removeTransaction(t))
        if gasFee != nil {
            total.Add(total, gasFee)
        }
    }
    return total
}

func execute(t *bean.Transaction) *big.Int {
    addr := accounts.PubKey2Address(t.Data.PubKey)
    nonce := t.Data.Nonce
    curN := database.GetNonce(addr, t.Data.ChainId)
    if curN + 1 != nonce {
        return nil
    }
    database.PutNonce(addr, t.Data.ChainId, curN + 1)
    gasPrice := big.NewInt(0).SetBytes(t.Data.GasPrice)
    gasLimit := big.NewInt(0).SetBytes(t.Data.GasLimit)
    gasFee := big.NewInt(0).Mul(gasLimit, gasPrice)
    balance := database.GetBalance(addr, t.Data.ChainId)
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
                    to := t.Data.To
                    balance2 := database.GetBalance(accounts.Hex2Address(to), t.Data.ChainId)
                    balance2.Add(balance2, amount)
                    database.PutBalance(accounts.Hex2Address(to), t.Data.ChainId, balance2)
                } else {
                    balance.Sub(balance, gasFee)
                }
            }
            database.PutBalance(addr, t.Data.ChainId, balance)
        }
    case MinerType:
        {
            // TODO if not the admin
            if gasLimit.Cmp(MinerGas) < 0 {
                balance.Sub(balance, gasFee)
            } else {
                balance.Sub(balance, gasFee)
                AddMiner(accounts.Bytes2Address(t.Data.Data))
            }
        }
    case CreateContractType:

    }
    return gasFee
}
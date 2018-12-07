package store

import (
    "math/big"
    "BlockChainTest/bean"
    "log"
    "fmt"
    "BlockChainTest/database"
    "BlockChainTest/accounts"
    "BlockChainTest/core"
    "BlockChainTest/core/vm"
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
    shift := new(big.Int).SetInt64(1000)
    ratio := new(big.Int).SetInt64(5)
    prize := new(big.Int).Mul(new(big.Int).Add(total, shift), ratio)
    leaderPrize := new(big.Int).Rsh(prize, 1)
    database.PutBalanceOutSideTransaction(accounts.PubKey2Address(b.Header.LeaderPubKey), b.Header.ChainId, leaderPrize)
    leftPrize := new(big.Int).Sub(prize, leaderPrize)
    minerNum := 0
    for _, elem := range b.Header.Bitmap {
        minerNum += elem
    }
    if minerNum == 0 {
        return total
    }
    minerPrize := new(big.Int).Div(leftPrize, new(big.Int).SetInt64(int64(minerNum)))
    for i, _ := range b.Header.Bitmap {
        if b.Header.Bitmap[i] == 1 {
            database.PutBalanceOutSideTransaction(accounts.PubKey2Address(b.Header.MinorPubKeys[i]), b.Header.ChainId, minerPrize)
        }
    }
    return total
}

func execute(t *bean.Transaction) *big.Int {
    fmt.Println("tttt: ", t.Data)
    addr := accounts.PubKey2Address(t.Data.PubKey)
    nonce := t.Data.Nonce
    curN := database.GetNonceOutsideTransaction(addr, t.Data.ChainId)
    if curN + 1 != nonce {
        return nil
    }
    database.PutNonceOutsideTransaction(addr, t.Data.ChainId, curN + 1)
    gasPrice := big.NewInt(0).SetBytes(t.Data.GasPrice)
    gasLimit := big.NewInt(0).SetBytes(t.Data.GasLimit)
    gasFee := big.NewInt(0).Mul(gasLimit, gasPrice)
    balance := database.GetBalanceOutsideTransaction(addr, t.Data.ChainId)
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
                    balance2 := database.GetBalanceOutsideTransaction(accounts.Hex2Address(to), t.Data.ChainId)
                    balance2.Add(balance2, amount)
                    database.PutBalanceOutSideTransaction(accounts.Hex2Address(to), t.Data.ChainId, balance2)
                } else {
                    balance.Sub(balance, gasFee)
                }
            }
            database.PutBalanceOutSideTransaction(addr, t.Data.ChainId, balance)
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
        {
            evm := vm.NewEVM()
            var gasFee = new(big.Int).Mul(CreateContractGas, gasPrice)
            fmt.Println("gas limit: ", gasLimit)
            fmt.Println("gas fee:   ", gasFee)
            fmt.Println("create contract gas: ", CreateContractGas)
            fmt.Println("balance: ", balance)
            if gasLimit.Cmp(CreateContractGas) < 0 {
                balance.Sub(balance, gasFee)
            } else {
                returnGas, _ := core.ApplyTransaction(evm, t)
                usedGas := new(big.Int).Sub(new(big.Int).SetBytes(t.Data.GasLimit), new(big.Int).SetUint64(returnGas))
                gasFee.Add(gasFee, new(big.Int).Mul(usedGas, gasPrice))
                balance.Sub(balance, gasFee)
                fmt.Println("returnGas: ", returnGas)
                fmt.Println("usedGas: ", usedGas)
                fmt.Println("gasFee: ", gasFee)
                fmt.Println("balance: ", balance)
            }
            evm.State.Commit()
            fmt.Println("db balance after commit: ", database.GetBalanceOutsideTransaction(addr, t.Data.ChainId))
            database.PutBalanceOutSideTransaction(addr, t.Data.ChainId, balance)
            fmt.Println("addr: ", addr.Hex())
            fmt.Println("t.Data.ChainId: ", t.Data.ChainId)
            fmt.Println("balance: ", balance)
            fmt.Println("db balance before commit: ", database.GetBalanceOutsideTransaction(addr, t.Data.ChainId))
        }
    case CallContractType:
        {
            evm := vm.NewEVM()
            var gasFee = new(big.Int).Mul(CallContractGas, gasPrice)
            if gasLimit.Cmp(CallContractGas) < 0 {
                balance.Sub(balance, gasFee)
            } else {
                returnGas, _ := core.ApplyTransaction(evm, t)
                usedGas := new(big.Int).Sub(new(big.Int).SetBytes(t.Data.GasLimit), new(big.Int).SetUint64(returnGas))
                gasFee.Add(gasFee, new(big.Int).Mul(usedGas, gasPrice))
                balance.Sub(balance, gasFee)
            }
            evm.State.Commit()
            database.PutBalanceOutSideTransaction(addr, t.Data.ChainId, balance)
        }
    }
    return gasFee
}
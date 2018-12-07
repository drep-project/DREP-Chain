package store

import (
    "fmt"
    "math/big"

    "BlockChainTest/log"
    "BlockChainTest/core"
    "BlockChainTest/bean"
    "BlockChainTest/database"
    "BlockChainTest/accounts"
    "BlockChainTest/core/vm"
    "encoding/json"
    "BlockChainTest/config"
    "bytes"
    "encoding/hex"
)

func ExecuteTransactions(b *bean.Block) *big.Int {
    if b == nil || b.Header == nil { // || b.Data == nil || b.Data.TxList == nil {
        log.Error("error block nil or header nil")
        return nil
    }
    dbTran := database.BeginTransaction()
    height := database.GetMaxHeightInsideTransaction(dbTran)
    if height + 1 != b.Header.Height {
        fmt.Println("error", height, b.Header.Height)
        return nil
    }
    total := big.NewInt(0)
    if b.Data == nil || b.Data.TxList == nil {
        return total
    }
    for _, t := range b.Data.TxList {
        _, gasFee := execute(dbTran, t)
        fmt.Println("Delete transaction ", *t)
        fmt.Println(removeTransaction(t))
        if gasFee != nil {
            total.Add(total, gasFee)
        }
    }

    prize := new(big.Int).Add(total, config.GetBlockPrize())

    leaderPrize := new(big.Int).Rsh(prize, 1)
    log.Trace("leader prize: ","Prize", leaderPrize)
    leaderAddr := accounts.PubKey2Address(b.Header.LeaderPubKey)
    balance := database.GetBalanceOutsideTransaction(leaderAddr, b.Header.ChainId)
    balance = new(big.Int).Add(balance, leaderPrize)
    database.PutBalanceOutSideTransaction(leaderAddr, b.Header.ChainId, balance)
    leftPrize := new(big.Int).Sub(prize, leaderPrize)
    minerNum := 0
    for _, elem := range b.MultiSig.Bitmap {
        if elem == 1 {
            minerNum++
        }
    }
    if minerNum == 0 {
        return total
    }
    minerPrize := new(big.Int).Div(leftPrize, new(big.Int).SetInt64(int64(minerNum)))
    for i, e := range b.MultiSig.Bitmap {
        if e == 1 {
            minerAddr := accounts.PubKey2Address(b.Header.MinorPubKeys[i])
            bal := database.GetBalanceOutsideTransaction(minerAddr, b.Header.ChainId)
            bal = new(big.Int).Add(bal, minerPrize)
            database.PutBalanceOutSideTransaction(minerAddr, b.Header.ChainId, bal)
        }
    }

    //stateRoot := database.GetStateRoot()
    //if bytes.Equal(b.Header.StateRoot, stateRoot) {
    //    fmt.Println("well ", len(b.Data.TxList))
    //    dbTran.Commit()
    //} else {
    //    fmt.Println("bad: ", len(b.Data.TxList))
    //    dbTran.Discard()
    //}
    dbTran.Commit()
    stateRoot := database.GetStateRoot()
    fmt.Println("state root 2: ", hex.EncodeToString(stateRoot))
    if bytes.Equal(b.Header.StateRoot, stateRoot) {
        fmt.Println()
        fmt.Println("matched ", hex.EncodeToString(b.Header.StateRoot), " vs ", hex.EncodeToString(stateRoot))
        fmt.Println()
        height++
        database.PutMaxHeightInsideTransaction(dbTran, height, GetChainId())
        database.PutBlockInsideTransaction(dbTran, b, GetChainId())
        dbTran.Commit()
    } else {
        fmt.Println()
        fmt.Println("not matched ", hex.EncodeToString(b.Header.StateRoot), " vs ", hex.EncodeToString(stateRoot))
        fmt.Println()
        dbTran.Discard()
    }
    return total
}

func execute(dbTran *database.Transaction, t *bean.Transaction) (gasUsed, gasFee *big.Int) {
    switch t.Data.Type {
    case TransferType:
        return executeTransferTransaction(dbTran, t)
    case MinerType:
        return executeMinerTransaction(dbTran, t)
    case CreateContractType:
        return executeCreateContractTransaction(dbTran, t)
    case CallContractType:
        return executeCallContractTransaction(dbTran, t)
    case CrossChainType:
        fmt.Println("execute crossed")
        return executeCrossChainTransaction(dbTran, t)
    }
    return nil, nil
}

func preExecute(dbTran *database.Transaction, t *bean.Transaction, gasWant *big.Int) (canExecute bool, addr accounts.CommonAddress,
    balance, gasPrice, gasUsed, gasFee *big.Int) {

    addr = accounts.PubKey2Address(t.Data.PubKey)
    gasUsed, gasFee = new(big.Int), new(big.Int)

    gasPrice = new(big.Int).SetBytes(t.Data.GasPrice)
    gasLimit := new(big.Int).SetBytes(t.Data.GasLimit)
    amountWant := new(big.Int).Mul(gasWant, gasPrice)
    balance = database.GetBalanceInsideTransaction(dbTran, addr, t.Data.ChainId)
    nonce := database.GetNonceInsideTransaction(dbTran, addr, t.Data.ChainId)

    if nonce + 1 != t.Data.Nonce || gasLimit.Cmp(gasWant) < 0 || balance.Cmp(amountWant) < 0 {
        gasUsed = new(big.Int).Set(gasLimit)
        gasFee = new(big.Int).Mul(gasLimit, gasPrice)
        balance = new(big.Int).Sub(balance, gasFee)
        if balance.Sign() < 0 {
            gasFee = new(big.Int).Add(balance, gasFee)
            balance = new(big.Int)
        }
        database.PutBalanceInsideTransaction(dbTran, addr, t.Data.ChainId, balance)
        return
    }

    canExecute = true
    gasUsed = new(big.Int).Set(gasWant)
    gasFee = new(big.Int).Set(amountWant)
    balance = new(big.Int).Sub(balance, gasFee)
    database.PutBalanceInsideTransaction(dbTran, addr, t.Data.ChainId, balance)
    database.PutNonceInsideTransaction(dbTran, addr, t.Data.ChainId, nonce + 1)
    return
}

func executeTransferTransaction(dbTran *database.Transaction, t *bean.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
    var (
        canExecute = false
        addr accounts.CommonAddress
        balance *big.Int
    )

    canExecute, addr, balance, _, gasUsed, gasFee = preExecute(dbTran, t, TransferGas)
    if !canExecute {
        return
    }

    amount := new(big.Int).SetBytes(t.Data.Amount)
    if balance.Cmp(amount) >= 0 {
        to := accounts.Hex2Address(t.Data.To)
        balance = new(big.Int).Sub(balance, amount)
        balance2 := database.GetBalanceInsideTransaction(dbTran, to, t.Data.ChainId)
        balance2 = new(big.Int).Add(balance2, amount)
        database.PutBalanceInsideTransaction(dbTran, addr, t.Data.ChainId, balance)
        database.PutBalanceInsideTransaction(dbTran, to, t.Data.DestChain, balance2)
    }
    return
}

func executeMinerTransaction(dbTran *database.Transaction, t *bean.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
    var (
        canExecute = false
        balance *big.Int
    )

    canExecute, _, balance, _, gasUsed, gasFee = preExecute(dbTran, t, MinerGas)
    if !canExecute {
        return
    }

    if balance.Sign() >= 0 {
        AddMiner(accounts.Bytes2Address(t.Data.Data))
    }
    return
}

func executeCreateContractTransaction(dbTran *database.Transaction, t *bean.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
    var (
        canExecute = false
        addr accounts.CommonAddress
        balance, gasPrice *big.Int
    )

    canExecute, addr, balance, gasPrice, gasUsed, gasFee = preExecute(dbTran, t, CreateContractGas)
    if !canExecute {
        return
    }

    evm := vm.NewEVM(dbTran)
    returnGas, _ := core.ApplyTransaction(evm, t)
    consumedGas := new(big.Int).Sub(new(big.Int).SetBytes(t.Data.GasLimit), new(big.Int).SetUint64(returnGas))
    consumedAmount := new(big.Int).Mul(consumedGas, gasPrice)
    if balance.Cmp(consumedAmount) >= 0 {
        gasUsed = new(big.Int).Add(gasUsed, consumedGas)
        gasFee = new(big.Int).Add(gasFee, consumedAmount)
        balance = new(big.Int).Sub(balance, consumedAmount)
        database.PutBalanceInsideTransaction(dbTran, addr, t.Data.ChainId, balance)
    }
    return
}

func executeCallContractTransaction(dbTran *database.Transaction, t *bean.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
    var (
        canExecute = false
        addr accounts.CommonAddress
        balance, gasPrice *big.Int
    )

    canExecute, addr, balance, gasPrice, gasUsed, gasFee = preExecute(dbTran, t, CallContractGas)
    if !canExecute {
        return
    }

    evm := vm.NewEVM(dbTran)
    returnGas, _ := core.ApplyTransaction(evm, t)
    consumedGas := new(big.Int).Sub(new(big.Int).SetBytes(t.Data.GasLimit), new(big.Int).SetUint64(returnGas))
    consumedAmount := new(big.Int).Mul(consumedGas, gasPrice)
    if balance.Cmp(consumedAmount) >= 0 {
        gasUsed = new(big.Int).Add(gasUsed, consumedGas)
        gasFee = new(big.Int).Add(gasFee, consumedAmount)
        balance = new(big.Int).Sub(balance, consumedAmount)
        database.PutBalanceInsideTransaction(dbTran, addr, t.Data.ChainId, balance)
    }
    return
}

func executeCrossChainTransaction(dbTran *database.Transaction, t *bean.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
    var (
        canExecute = false
        addr accounts.CommonAddress
        balance, gasPrice *big.Int
    )

    canExecute, addr, balance, gasPrice, gasUsed, gasFee = preExecute(dbTran, t, CrossChainGas)
    if !canExecute {
        return
    }

    sumGas := new(big.Int)
    data := t.Data.Data
    var trans []*bean.Transaction
    err := json.Unmarshal(data, &trans)
    if err == nil {
        for _, tx := range trans {
            if tx.Data.Type == CrossChainType {
                continue
            }
            g, _ := execute(dbTran, tx)
            sumGas = new(big.Int).Add(sumGas, g)
        }
    }

    sumAmount := new(big.Int).Mul(sumGas, gasPrice)
    if balance.Cmp(sumAmount) >= 0 {
        gasUsed = new(big.Int).Add(gasUsed, sumGas)
        gasFee = new(big.Int).Add(gasFee, sumAmount)
        balance = new(big.Int).Sub(balance, sumAmount)
        database.PutBalanceInsideTransaction(dbTran, addr, t.Data.ChainId, balance)
    }
    return
}

//TODO
//局部回滚
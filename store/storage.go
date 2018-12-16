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
    dt := database.BeginTransaction()
    height := database.GetMaxHeight()
    if height + 1 != b.Header.Height {
        fmt.Println("error", height, b.Header.Height)
        return nil
    }
    total := big.NewInt(0)
    if b.Data == nil || b.Data.TxList == nil {
        return total
    }
    for _, t := range b.Data.TxList {
        subDt := dt.BeginTransaction()
        _, gasFee := execute(subDt, t)
        fmt.Println("Delete transaction ", *t)
        fmt.Println(removeTransaction(t))
        if gasFee != nil {
            total.Add(total, gasFee)
        }
        subDt.Commit()
    }

    stateRoot := dt.GetTotalStateRoot()
    if bytes.Equal(b.Header.StateRoot, stateRoot) {
        fmt.Println()
        fmt.Println("matched ", hex.EncodeToString(b.Header.StateRoot), " vs ", hex.EncodeToString(stateRoot))
        height++
        database.PutMaxHeight(height)
        database.PutBlock(b)
        fmt.Println("received block: ", b.Header, " ", b.Data, " ", b.MultiSig)
        dt.Commit()
        //fmt.Println("11111: ", database.GetBalance(GetAddress(), GetChainId()))
        distributeBlockPrize(b, total)
        //fmt.Println("22222: ", database.GetBalance(GetAddress(), GetChainId()))
    } else {
        fmt.Println("not matched ", hex.EncodeToString(b.Header.StateRoot), " vs ", hex.EncodeToString(stateRoot))
        dt.Discard()
    }
    return total
}

func execute(dt database.Transactional, t *bean.Transaction) (gasUsed, gasFee *big.Int) {
    switch t.Data.Type {
    case TransferType:
       return executeTransferTransaction(dt, t)
    case MinerType:
       return executeMinerTransaction(dt, t)
    case CreateContractType:
       return executeCreateContractTransaction(dt, t)
    case CallContractType:
       return executeCallContractTransaction(dt, t)
    case CrossChainType:
       return executeCrossChainTransaction(dt, t)
    }
    return nil, nil
}

func canExecute(dt database.Transactional, t *bean.Transaction, gasFloor, gasCap *big.Int) (canExecute bool, addr accounts.CommonAddress, balance, gasLimit, gasPrice *big.Int) {
    addr = accounts.PubKey2Address(t.Data.PubKey)
    balance = database.GetBalance(addr, t.Data.ChainId)
    nonce := database.GetNonce(addr, t.Data.ChainId) + 1
    gasLimit = new(big.Int).SetBytes(t.Data.GasLimit)
    gasPrice = new(big.Int).SetBytes(t.Data.GasPrice)
    subDt := dt.BeginTransaction()
    database.PutNonce(subDt, addr, t.Data.ChainId, nonce)
    subDt.Commit()

    if nonce != t.Data.Nonce {
        return
    }
    if gasFloor != nil {
        amountFloor := new(big.Int).Mul(gasFloor, gasPrice)
        if gasLimit.Cmp(gasFloor) < 0 || amountFloor.Cmp(balance) > 0 {
            return
        }
    }
    if gasCap != nil {
        amountCap := new(big.Int).Mul(gasCap, gasPrice)
        if amountCap.Cmp(balance) > 0 {
            return
        }
    }

    canExecute = true
    return
}

func deduct(dt database.Transactional, addr accounts.CommonAddress, chainId config.ChainIdType, balance, gasFee *big.Int) (leftBalance, actualFee *big.Int) {
    subDt := dt.BeginTransaction()
    leftBalance = new(big.Int).Sub(balance, gasFee)
    actualFee = new(big.Int).Set(gasFee)
    if leftBalance.Sign() < 0 {
        actualFee = new(big.Int).Set(balance)
        leftBalance = new(big.Int)
    }
    database.PutBalance(subDt, addr, chainId, leftBalance)
    subDt.Commit()
    return leftBalance, actualFee
}

func preExecute(dt database.Transactional, t *bean.Transaction, gasWant *big.Int) (canExecute bool, addr accounts.CommonAddress,
   balance, gasPrice, gasUsed, gasFee *big.Int) {

   addr = accounts.PubKey2Address(t.Data.PubKey)
   gasUsed, gasFee = new(big.Int), new(big.Int)

   gasPrice = new(big.Int).SetBytes(t.Data.GasPrice)
   gasLimit := new(big.Int).SetBytes(t.Data.GasLimit)
   amountWant := new(big.Int).Mul(gasWant, gasPrice)
   balance = database.GetBalance(addr, t.Data.ChainId)
   nonce := database.GetNonce(addr, t.Data.ChainId)

   if nonce + 1 != t.Data.Nonce || gasLimit.Cmp(gasWant) < 0 || balance.Cmp(amountWant) < 0 {
       gasUsed = new(big.Int).Set(gasLimit)
       gasFee = new(big.Int).Mul(gasLimit, gasPrice)
       balance = new(big.Int).Sub(balance, gasFee)
       if balance.Sign() < 0 {
           gasFee = new(big.Int).Add(balance, gasFee)
           balance = new(big.Int)
       }
       database.PutBalance(dt, addr, t.Data.ChainId, balance)
       return
   }

   canExecute = true
   gasUsed = new(big.Int).Set(gasWant)
   gasFee = new(big.Int).Set(amountWant)
   balance = new(big.Int).Sub(balance, gasFee)
   database.PutBalance(dt, addr, t.Data.ChainId, balance)
   database.PutNonce(dt, addr, t.Data.ChainId, nonce + 1)
   return
}

func executeTransferTransaction(dt database.Transactional, t *bean.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
    var (
       can bool
       addr accounts.CommonAddress
       balance, gasPrice *big.Int
    )

    gasUsed, gasFee = new(big.Int), new(big.Int)
    subDt := dt.BeginTransaction()
    can, addr, balance, _, gasPrice = canExecute(subDt, t, TransferGas, nil)
    if !can {
       return
    }

    gasUsed = new(big.Int).Set(TransferGas)
    gasFee = new(big.Int).Mul(gasUsed, gasPrice)
    balance, gasFee = deduct(subDt, addr, t.Data.ChainId, balance, gasFee)
    amount := new(big.Int).SetBytes(t.Data.Amount)
    if balance.Cmp(amount) >= 0 {
       to := accounts.Hex2Address(t.Data.To)
       balance = new(big.Int).Sub(balance, amount)
       balanceTo := database.GetBalance(to, t.Data.DestChain)
       balanceTo = new(big.Int).Add(balanceTo, amount)
       database.PutBalance(subDt, addr, t.Data.ChainId, balance)
       database.PutBalance(subDt, to, t.Data.DestChain, balanceTo)
       subDt.Commit()
    }
    return
}

func executeMinerTransaction(dt database.Transactional, t *bean.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
    var (
       can bool
       addr accounts.CommonAddress
       balance, gasPrice *big.Int
    )

    subDt := dt.BeginTransaction()
    gasUsed, gasFee = new(big.Int), new(big.Int)
    can, addr, balance, _, gasPrice = canExecute(subDt, t, MinerGas, nil)
    if !can {
       return
    }

    gasUsed = new(big.Int).Set(MinerGas)
    gasFee = new(big.Int).Mul(gasUsed, gasPrice)
    balance, gasFee = deduct(subDt, addr, t.Data.ChainId, balance, gasFee)
    if balance.Sign() >= 0 {
       AddMiner(accounts.Bytes2Address(t.Data.Data))
    }
    return
}

func executeCreateContractTransaction(dt database.Transactional, t *bean.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
    var (
       can bool
       addr accounts.CommonAddress
       balance, gasLimit, gasPrice *big.Int
    )

    subDt := dt.BeginTransaction()
    gasUsed, gasFee = new(big.Int), new(big.Int)
    can, addr, _, gasLimit, gasPrice = canExecute(subDt, t, nil, CreateContractGas)
    if !can {
       return
    }

    evm := vm.NewEVM(subDt)
    returnGas, _ := core.ApplyTransaction(evm, t)
    gasUsed = new(big.Int).Sub(gasLimit, new(big.Int).SetUint64(returnGas))
    gasFee = new(big.Int).Mul(gasUsed, gasPrice)
    balance = database.GetBalance(addr, t.Data.ChainId)
    _, gasFee = deduct(subDt, addr, t.Data.ChainId, balance, gasFee)
    subDt.Commit()
    return
}

func executeCallContractTransaction(dt database.Transactional, t *bean.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
    var (
        can bool
        addr accounts.CommonAddress
        balance, gasLimit, gasPrice *big.Int
    )

    gasUsed, gasFee = new(big.Int), new(big.Int)
    subDt := dt.BeginTransaction()
    can, addr, _, gasLimit, gasPrice = canExecute(subDt, t,nil, CallContractGas)
    if !can {
        return
    }

    evm := vm.NewEVM(subDt)
    returnGas, _ := core.ApplyTransaction(evm, t)
    gasUsed = new(big.Int).Sub(gasLimit, new(big.Int).SetUint64(returnGas))
    gasFee = new(big.Int).Mul(gasUsed, gasPrice)
    balance = database.GetBalance(addr, t.Data.ChainId)
    _, gasFee = deduct(subDt, addr, t.Data.ChainId, balance, gasFee)
    subDt.Commit()
    return
}

func executeCrossChainTransaction(dt database.Transactional, t *bean.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
    var (
        can bool
        addr accounts.CommonAddress
        balance, gasPrice *big.Int
    )

    gasUsed, gasFee = new(big.Int), new(big.Int)
    subDt := dt.BeginTransaction()
    can, addr,  _, _, gasPrice = canExecute(subDt, t, nil, CrossChainGas)
    if !can {
        return new(big.Int), new(big.Int)
    }

    cct := &bean.CrossChainTransaction{}
    err := json.Unmarshal(t.Data.Data, &cct)
    if err != nil {
       return new(big.Int), new(big.Int)
    }

    gasSum := new(big.Int)
    for _, tx := range cct.Trans {
       if tx.Data.Type == CrossChainType {
           continue
       }
       g, _ := execute(subDt, tx)
       gasSum = new(big.Int).Add(gasSum, g)
    }

    if !bytes.Equal(subDt.GetChainStateRoot(cct.ChainId), cct.StateRoot) {
       subDt.Discard()
    } else {
        amountSum := new(big.Int).Mul(gasSum, gasPrice)
        balance = database.GetBalance(addr, t.Data.ChainId)
        if balance.Cmp(amountSum) >= 0 {
            gasUsed = new(big.Int).Set(gasSum)
            gasFee = new(big.Int).Set(amountSum)
            _, gasFee = deduct(subDt, addr, t.Data.ChainId, balance, gasFee)
            subDt.Commit()
        } else {
            subDt.Discard()
        }
    }
    return
}

func distributeBlockPrize(b *bean.Block, total *big.Int) {
    dt := database.BeginTransaction()
    str := config.GetConfig().Blockprize.String()
    val := new (big.Int)
    val.SetString(str,10)
    prize := new(big.Int).Add(total, val)
    if b.Header.Height > 2 {
        prize = new(big.Int)
    }
    leaderPrize := new(big.Int).Rsh(prize, 1)
    fmt.Println("leader prize: ", leaderPrize)
    leaderAddr := accounts.PubKey2Address(b.Header.LeaderPubKey)
    balance := database.GetBalance(leaderAddr, b.Header.ChainId)
    balance = new(big.Int).Add(balance, leaderPrize)
    database.PutBalance(dt, leaderAddr, b.Header.ChainId, balance)
    leftPrize := new(big.Int).Sub(prize, leaderPrize)
    minerNum := 0
    for _, elem := range b.MultiSig.Bitmap {
        if elem == 1 {
            minerNum++
        }
    }
    if minerNum == 0 {
        dt.Commit()
        return
    }
    minerPrize := new(big.Int).Div(leftPrize, new(big.Int).SetInt64(int64(minerNum)))
    for i, e := range b.MultiSig.Bitmap {
        if e == 1 {
            minerAddr := accounts.PubKey2Address(b.Header.MinorPubKeys[i])
            bal := database.GetBalance(minerAddr, b.Header.ChainId)
            bal = new(big.Int).Add(bal, minerPrize)
            database.PutBalance(dt, minerAddr, b.Header.ChainId, bal)
        }
    }
    dt.Commit()
    return
}
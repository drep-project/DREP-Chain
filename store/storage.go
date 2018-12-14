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
        distributeBlockPrize(b, total)
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
       canExecute = false
       addr accounts.CommonAddress
       balance *big.Int
   )

   subDt := dt.BeginTransaction()
   canExecute, addr, balance, _, gasUsed, gasFee = preExecute(subDt, t, TransferGas)
   if !canExecute {
       subDt.Commit()
       return
   }

   amount := new(big.Int).SetBytes(t.Data.Amount)
   if balance.Cmp(amount) >= 0 {
       to := accounts.Hex2Address(t.Data.To)
       balance = new(big.Int).Sub(balance, amount)
       balance2 := database.GetBalance(to, t.Data.DestChain)
       balance2 = new(big.Int).Add(balance2, amount)
       database.PutBalance(subDt, addr, t.Data.ChainId, balance)
       database.PutBalance(subDt, to, t.Data.DestChain, balance2)
       subDt.Commit()
   }

   dt.Commit()
   return
}

func executeMinerTransaction(dt database.Transactional, t *bean.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
   var (
       canExecute = false
       balance *big.Int
   )

   subDt := dt.BeginTransaction()
   canExecute, _, balance, _, gasUsed, gasFee = preExecute(subDt, t, MinerGas)
   if !canExecute {
       subDt.Commit()
       return
   }

   if balance.Sign() >= 0 {
       AddMiner(accounts.Bytes2Address(t.Data.Data))
   }

   dt.Commit()
   return
}

func executeCreateContractTransaction(dt database.Transactional, t *bean.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
   var (
       canExecute = false
       addr accounts.CommonAddress
       balance, gasPrice *big.Int
   )

   subDt := dt.BeginTransaction()
   canExecute, addr, _, gasPrice, gasUsed, gasFee = preExecute(subDt, t, CreateContractGas)
   if !canExecute {
       subDt.Commit()
       return
   }

   evm := vm.NewEVM(subDt)
   returnGas, err := core.ApplyTransaction(evm, t)
   consumedGas := new(big.Int).Sub(new(big.Int).SetBytes(t.Data.GasLimit), new(big.Int).SetUint64(returnGas))
   consumedAmount := new(big.Int).Mul(consumedGas, gasPrice)
   if err != nil && balance.Cmp(consumedAmount) >= 0 {
       gasUsed = new(big.Int).Add(gasUsed, consumedGas)
       gasFee = new(big.Int).Add(gasFee, consumedAmount)
       balance = database.GetBalance(addr, t.Data.ChainId)
       balance = new(big.Int).Sub(balance, consumedAmount)
       database.PutBalance(subDt, addr, t.Data.ChainId, balance)
       subDt.Commit()
   } else {
       subDt.Discard()
   }
   dt.Commit()
   return
}

func executeCallContractTransaction(dt database.Transactional, t *bean.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
   var (
       canExecute = false
       addr accounts.CommonAddress
       balance, gasPrice *big.Int
   )

   subDt := dt.BeginTransaction()
   canExecute, addr, _, gasPrice, gasUsed, gasFee = preExecute(dt, t, CallContractGas)
   if !canExecute {
       subDt.Commit()
       return
   }

   evm := vm.NewEVM(subDt)
   returnGas, err := core.ApplyTransaction(evm, t)
   consumedGas := new(big.Int).Sub(new(big.Int).SetBytes(t.Data.GasLimit), new(big.Int).SetUint64(returnGas))
   consumedAmount := new(big.Int).Mul(consumedGas, gasPrice)
   balance = database.GetBalance(addr, t.Data.ChainId)
   if err != nil && balance.Cmp(consumedAmount) >= 0 {
       gasUsed = new(big.Int).Add(gasUsed, consumedGas)
       gasFee = new(big.Int).Add(gasFee, consumedAmount)
       balance = new(big.Int).Sub(balance, consumedAmount)
       database.PutBalance(subDt, addr, t.Data.ChainId, balance)
       subDt.Commit()
   } else {
       subDt.Discard()
   }
   dt.Commit()
   return
}

func executeCrossChainTransaction(dt database.Transactional, t *bean.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
   var (
       canExecute = false
       addr accounts.CommonAddress
       gasPrice *big.Int
   )

   subDt := dt.BeginTransaction()
   canExecute, addr, _, gasPrice, gasUsed, gasFee = preExecute(dt, t, CrossChainGas)
   if !canExecute {
       subDt.Commit()
       return
   }

   cct := &bean.CrossChainTransaction{}
   err := json.Unmarshal(t.Data.Data, cct)
   if err != nil {
       subDt.Commit()
       return
   }

   sumGas := new(big.Int)
   for _, tx := range cct.Trans {
       if tx.Data.Type == CrossChainType {
           continue
       }
       g, _ := execute(subDt, tx)
       sumGas = new(big.Int).Add(sumGas, g)
   }
   if !bytes.Equal(subDt.GetChainStateRoot(cct.ChainId), cct.StateRoot) {
       subDt.Discard()
   }

   sumAmount := new(big.Int).Mul(sumGas, gasPrice)
   balance := database.GetBalance(addr, t.Data.ChainId)
   if balance.Cmp(sumAmount) >= 0 {
       gasUsed = new(big.Int).Add(gasUsed, sumGas)
       gasFee = new(big.Int).Add(gasFee, sumAmount)
       balance = new(big.Int).Sub(balance, sumAmount)
       database.PutBalance(subDt, addr, t.Data.ChainId, balance)
       subDt.Commit()
   } else {
       subDt.Discard()
   }
   dt.Commit()
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
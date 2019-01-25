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
    "BlockChainTest/config"
    "BlockChainTest/mycrypto"
    "encoding/json"
    "BlockChainTest/repjs"
    "strconv"
    "BlockChainTest/wasm"
)

var (
    lastLeader *mycrypto.Point
    lastMinors []*mycrypto.Point
    lastPrize  *big.Int
    lastIndex  = 0
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
    total := big.NewInt(0)
    if b.Data == nil || b.Data.TxList == nil {
        return total
    }
    for _, t := range b.Data.TxList {
        if t.Data.Type == BlockPrizeType {
            executeBlockPrizeTransaction(t)
            continue
        }
        if t.Data.Type == GainType {
            executeGainTransaction(t)
            continue
        }
        gasFee := execute(t)
        fmt.Println("Delete transaction ", *t)
        fmt.Println(removeTransaction(t))
        if gasFee != nil {
            total.Add(total, gasFee)
        }
    }
    savePrizeInfo(b, total)
    database.PutMaxHeight(b.Header.Height)
    database.PutBlock(b)
    database.PutLastTimestamp(b.Header.Timestamp)
    th, _ := json.Marshal(b.Header)
    database.PutPreviousHash(mycrypto.Hash256(th))

    Liquidate(b.Header.Height, int(b.Header.Height))
    return total
}

func execute(t *bean.Transaction) *big.Int {
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
            if gasLimit.Cmp(CreateContractGas) < 0 {
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

func executeWithT(t *bean.Transaction, dbTran *database.Transaction) *big.Int {
    addr := accounts.PubKey2Address(t.Data.PubKey)
    nonce := t.Data.Nonce
    curN := database.GetNonceInsideTransaction(dbTran, addr, t.Data.ChainId)
    if curN + 1 != nonce {
        return nil
    }
    database.PutNonceInsideTransaction(dbTran, addr, t.Data.ChainId, curN + 1)
    gasPrice := big.NewInt(0).SetBytes(t.Data.GasPrice)
    gasLimit := big.NewInt(0).SetBytes(t.Data.GasLimit)
    gasFee := big.NewInt(0).Mul(gasLimit, gasPrice)
    balance := database.GetBalanceInsideTransaction(dbTran, addr, t.Data.ChainId)
    if gasFee.Cmp(balance) > 0 {
        log.Fatal("Error, gas not right")
        return nil
    }
    switch t.Data.Type {
    case TransferType:
        {
            if gasLimit.Cmp(TransferGas) < 0 {
                balance.Sub(balance, gasFee)
                return gasLimit
            } else {
                amount := big.NewInt(0).SetBytes(t.Data.Amount)
                total := big.NewInt(0).Add(amount, gasFee)
                if balance.Cmp(total) >= 0 {
                    balance.Sub(balance, total)
                    to := t.Data.To
                    balance2 := database.GetBalanceInsideTransaction(dbTran, accounts.Hex2Address(to), t.Data.ChainId)
                    balance2.Add(balance2, amount)
                    database.PutBalanceInsideTransaction(dbTran, accounts.Hex2Address(to), t.Data.ChainId, balance2)
                } else {
                    balance.Sub(balance, gasFee)
                }
            }
            database.PutBalanceInsideTransaction(dbTran, addr, t.Data.ChainId, balance)
            return TransferGas
        }
    case MinerType:
        {
            // TODO if not the admin
            if gasLimit.Cmp(MinerGas) < 0 {
                balance.Sub(balance, gasFee)
                return gasLimit
            } else {
                balance.Sub(balance, gasFee)
                AddMiner(accounts.Bytes2Address(t.Data.Data))
                return MinerGas
            }
        }
    case CreateContractType:
        {
            evm := vm.NewEVM()
            var gasFee = new(big.Int).Mul(CreateContractGas, gasPrice)
            if gasLimit.Cmp(CreateContractGas) < 0 {
                balance.Sub(balance, gasFee)
                return gasLimit
            } else {
                returnGas, _ := core.ApplyTransaction(evm, t)
                usedGas := new(big.Int).Sub(new(big.Int).SetBytes(t.Data.GasLimit), new(big.Int).SetUint64(returnGas))
                gasFee.Add(gasFee, new(big.Int).Mul(usedGas, gasPrice))
                balance.Sub(balance, gasFee)
            }
            evm.State.Commit()
            database.PutBalanceInsideTransaction(dbTran, addr, t.Data.ChainId, balance)
            return CreateContractGas
        }
    case CallContractType:
        {
            evm := vm.NewEVM()
            var gasFee = new(big.Int).Mul(CallContractGas, gasPrice)
            if gasLimit.Cmp(CallContractGas) < 0 {
                balance.Sub(balance, gasFee)
                return gasLimit
            } else {
                returnGas, _ := core.ApplyTransaction(evm, t)
                usedGas := new(big.Int).Sub(new(big.Int).SetBytes(t.Data.GasLimit), new(big.Int).SetUint64(returnGas))
                gasFee.Add(gasFee, new(big.Int).Mul(usedGas, gasPrice))
                balance.Sub(balance, gasFee)
            }
            evm.State.Commit()
            database.PutBalanceInsideTransaction(dbTran, addr, t.Data.ChainId, balance)
            return CallContractGas
        }
    }
    return new(big.Int)
}

func executeBlockPrizeTransaction(t *bean.Transaction) {
    addr := accounts.Hex2Address(t.Data.To)
    balance := database.GetBalanceOutsideTransaction(addr, t.Data.DestChain)
    balance = new(big.Int).Add(balance, new(big.Int).SetBytes(t.Data.Amount))
    database.PutBalanceOutSideTransaction(addr, t.Data.DestChain, balance)
    return
}

func savePrizeInfo(block *bean.Block, total *big.Int) {
    fmt.Println()
    fmt.Println("block gas used: ", new(big.Int).SetBytes(block.Header.GasUsed))
    fmt.Println()
    lastLeader = block.Header.LeaderPubKey
    lastMinors = block.Header.MinorPubKeys
    basePrize := config.GetBlockPrize()
    lastPrize = new(big.Int).Add(basePrize, total)
    lastIndex++
    if lastIndex > 6 {
        lastIndex = 0
    }
}

func executeGainTransaction(t *bean.Transaction) {
    var records []map[string] interface{}
    err := json.Unmarshal(t.Data.Data, &records)
    if err != nil {
        return
    }
    increments := make([]map[string] interface{}, len(records))
    platformID := strconv.FormatInt(GetChainId(), 10)
    for i, r := range records {
        uid := r["Addr"].(string)
        ret := repjs.GetProfile(platformID, uid)
        if ret == nil {
            return
        }
        repID := ret["RepID"].(string)
        groupID := ret["GroupID"].(uint64)
        tracer := database.GetTracer(platformID, repID)
        if tracer == nil {
            repjs.RegisterUser(platformID, repID, groupID)

            wasm.RegisterUser()

        }
        increments[i] = make(map[string] interface{})
        increments[i]["RepID"] = repID
        increments[i]["Day"] = r["Day"]
        increments[i]["Gain"] = r["Gain"]
    }
    repjs.AddGain(platformID, increments)
}

func Liquidate(height int64, until int) {
    platformID := strconv.FormatInt(GetChainId(), 10)
    groupID := uint64(height % 5)
    repjs.LiquidateRepByGroup(platformID, groupID, until)
}
package service

import (
    "errors"
    "fmt"
    chainComponent "github.com/drep-project/drep-chain/chain/component"
    "github.com/drep-project/drep-chain/chain/component/vm"
    chainTypes "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/drep-chain/common"
    "github.com/drep-project/drep-chain/crypto"
    "github.com/drep-project/drep-chain/log"
    "math/big"

    "bytes"
    "encoding/hex"
    "encoding/json"
    "net/http"
    "net/url"
    "strconv"
)

var (
    childTrans []*chainTypes.Transaction
)

func (chainService *ChainService) ExecuteTransactions(b *chainTypes.Block) (*big.Int, error) {
    if b == nil || b.Header == nil { // || b.Data == nil || b.Data.TxList == nil {
        return nil, errors.New("error block nil or header nil")
    }
    height := chainService.DatabaseService.GetMaxHeight()
    if height + 1 != b.Header.Height {
        msg := fmt.Sprintf("not corrent height CurrentHeight: %d, ReceiveHeight: %d", height, b.Header.Height)
        log.Error(msg)
        return nil, errors.New(msg)
    }

    chainService.DatabaseService.BeginTransaction()
    total := big.NewInt(0)
    if b.Data == nil || b.Data.TxList == nil {
        return total, nil
    }
    for _, t := range b.Data.TxList {
        _, gasFee := chainService.execute(t)
        if gasFee != nil {
            total.Add(total, gasFee)
        }
    }

    stateRoot := chainService.DatabaseService.GetStateRoot()
    if bytes.Equal(b.Header.StateRoot, stateRoot) {
        fmt.Println()
        fmt.Println("matched ", hex.EncodeToString(b.Header.StateRoot), " vs ", hex.EncodeToString(stateRoot))
        height++
        chainService.DatabaseService.PutMaxHeight(height)
        chainService.DatabaseService.PutBlock(b)
        chainService.DatabaseService.Commit()
        fmt.Println("received block: ", true)
        fmt.Println()

        chainService.accumulateRewards(b, chainService.ChainID())
        chainService.preSync(b)
        chainService.doSync(height)
    } else {
        chainService.DatabaseService.Discard()
        fmt.Println()
        fmt.Println("not matched ", hex.EncodeToString(b.Header.StateRoot), " vs ", hex.EncodeToString(stateRoot))
        fmt.Println("received block: ", false)
        fmt.Println()
    }
    return total, nil
}


func (chainService *ChainService) preSync(block *chainTypes.Block) {
    if !chainService.isRelay && chainService.chainId != chainService.RootChain() {
        return
    }
    if childTrans == nil {
        childTrans = make([]*chainTypes.Transaction, 0)
    }
    childTrans = append(childTrans, block.Data.TxList...)
}

func (chainService *ChainService) doSync(height int64) {
    if !chainService.isRelay || chainService.chainId == chainService.RootChain() || height % 2 != 0 || height == 0 {
        return
    }
    cct := &chainTypes.CrossChainTransaction{
        ChainId: chainService.chainId,
        StateRoot: chainService.DatabaseService.GetStateRoot(),
        Trans: childTrans,
    }
    data, err := json.Marshal(cct)
    if err != nil {
        return
    }
    values := url.Values{}
    values.Add("data", string(data))
    body := values.Encode()
    urlStr := "http://localhost:" + strconv.Itoa(chainService.Config.RemotePort) + "/SyncChildChain?" + body
    http.Get(urlStr)
    childTrans = nil
}

func (chainService *ChainService) execute(t *chainTypes.Transaction) (gasUsed, gasFee *big.Int) {
    switch t.Data.Type {
    case TransferType:
       return chainService.executeTransferTransaction(t)
    case CreateContractType:
       return chainService.executeCreateContractTransaction(t)
    case CallContractType:
       return chainService.executeCallContractTransaction(t)
    //case CrossChainType:
    //   return chainService.executeCrossChainTransaction(t)
    }
    return nil, nil
}

func (chainService *ChainService) canExecute(t *chainTypes.Transaction, gasFloor, gasCap *big.Int) (canExecute bool, addr crypto.CommonAddress, balance, gasLimit, gasPrice *big.Int) {
    addr = crypto.PubKey2Address(t.Data.PubKey)
    balance = chainService.DatabaseService.GetBalance(addr, t.Data.ChainId, true)
    nonce :=  chainService.DatabaseService.GetNonce(addr,true) + 1
    chainService.DatabaseService.PutNonce(addr, nonce,true)

    if nonce != t.Data.Nonce {
        return
    }
    if gasFloor != nil {
        amountFloor := new(big.Int).Mul(gasFloor, t.Data.GasPrice)
        if t.Data.GasLimit.Cmp(gasFloor) < 0 || amountFloor.Cmp(balance) > 0 {
            return
        }
    }
    if gasCap != nil {
        amountCap := new(big.Int).Mul(gasCap, t.Data.GasPrice)
        if amountCap.Cmp(balance) > 0 {
            return
        }
    }

    canExecute = true
    return
}

func (chainService *ChainService) deduct(addr crypto.CommonAddress, chainId common.ChainIdType, balance, gasFee *big.Int) (leftBalance, actualFee *big.Int) {
    leftBalance = new(big.Int).Sub(balance, gasFee)
    actualFee = new(big.Int).Set(gasFee)
    if leftBalance.Sign() < 0 {
        actualFee = new(big.Int).Set(balance)
        leftBalance = new(big.Int)
    }
    chainService.DatabaseService.PutBalance(addr, chainId, leftBalance, true)
    return leftBalance, actualFee
}

func (chainService *ChainService) executeTransferTransaction(t *chainTypes.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
    var (
       can bool
       addr crypto.CommonAddress
       balance, gasPrice *big.Int
    )

    gasUsed, gasFee = new(big.Int), new(big.Int)
    can, addr, balance, _, gasPrice = chainService.canExecute(t, TransferGas, nil)
    if !can {
       return
    }

    gasUsed = new(big.Int).Set(TransferGas)
    gasFee = new(big.Int).Mul(gasUsed, gasPrice)
    balance, gasFee = chainService.deduct(addr, t.Data.ChainId, balance, gasFee)
    if balance.Cmp(t.Data.Amount) >= 0 {
       balance = new(big.Int).Sub(balance, t.Data.Amount)
       balanceTo := chainService.DatabaseService.GetBalance(t.Data.To, t.Data.ChainId, true)
       balanceTo = new(big.Int).Add(balanceTo, t.Data.Amount)
       chainService.DatabaseService.PutBalance(addr, t.Data.ChainId, balance, true)
       chainService.DatabaseService.PutBalance(t.Data.To, t.Data.ChainId, balanceTo, true)
    }
    return
}

func (chainService *ChainService) executeCreateContractTransaction(t *chainTypes.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
    var (
       can bool
       addr crypto.CommonAddress
       balance, gasLimit, gasPrice *big.Int
    )
    gasUsed, gasFee = new(big.Int), new(big.Int)
    can, addr, _, gasLimit, gasPrice = chainService.canExecute(t, nil, CreateContractGas)
    if !can {
       return
    }

    evm := vm.NewEVM(chainService.DatabaseService)
    returnGas, _ := chainComponent.ApplyTransaction(evm, t)
    gasUsed = new(big.Int).Sub(gasLimit, new(big.Int).SetUint64(returnGas))
    gasFee = new(big.Int).Mul(gasUsed, gasPrice)
    balance = chainService.DatabaseService.GetBalance(addr, t.Data.ChainId, true)
    _, gasFee = chainService.deduct(addr, t.Data.ChainId, balance, gasFee)
    return
}

func (chainService *ChainService) executeCallContractTransaction(t *chainTypes.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
    var (
        can bool
        addr crypto.CommonAddress
        balance, gasLimit, gasPrice *big.Int
    )

    gasUsed, gasFee = new(big.Int), new(big.Int)
    can, addr, _, gasLimit, gasPrice = chainService.canExecute(t,nil, CallContractGas)
    if !can {
        return
    }

    evm := vm.NewEVM(chainService.DatabaseService)
    returnGas, _ := chainComponent.ApplyTransaction(evm, t)
    gasUsed = new(big.Int).Sub(gasLimit, new(big.Int).SetUint64(returnGas))
    gasFee = new(big.Int).Mul(gasUsed, gasPrice)
    balance = chainService.DatabaseService.GetBalance(addr, t.Data.ChainId, true)
    _, gasFee = chainService.deduct(addr, t.Data.ChainId, balance, gasFee)
    return
}

//func (chainService *ChainService) executeCrossChainTransaction(t *chainTypes.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
//    var (
//        can bool
//        addr crypto.CommonAddress
//        balance, gasPrice *big.Int
//    )
//
//    gasUsed, gasFee = new(big.Int), new(big.Int)
//    can, addr,  _, _, gasPrice = chainService.canExecute(t, nil, CrossChainGas)
//    if !can {
//        return new(big.Int), new(big.Int)
//    }
//
//    cct := &chainTypes.CrossChainTransaction{}
//    err := json.Unmarshal(t.Data.Data, cct)
//    if err != nil {
//        fmt.Println("err: ", err)
//        return new(big.Int), new(big.Int)
//    }
//
//    gasSum := new(big.Int)
//    for _, tx := range cct.Trans {
//       if tx.Data.Type == CrossChainType {
//           continue
//       }
//       g, _ := chainService.execute(tx)
//       gasSum = new(big.Int).Add(gasSum, g)
//    }
//
//    if !bytes.Equal(chainService.databaseService.GetStateRoot(), cct.StateRoot) {
//       //subDt.Discard()
//    } else {
//        amountSum := new(big.Int).Mul(gasSum, gasPrice)
//        balance = chainService.databaseService.GetBalance(addr, t.Data.ChainId, true)
//        if balance.Cmp(amountSum) >= 0 {
//            gasUsed = new(big.Int).Set(gasSum)
//            gasFee = new(big.Int).Set(amountSum)
//            _, gasFee = chainService.deduct(addr, t.Data.ChainId, balance, gasFee)
//            //subDt.Commit()
//        } else {
//            //subDt.Discard()
//        }
//    }
//    return
//}

//func preExecuteCrossChainTransaction(dt database.Transactional, t *chainTypes.Transaction) (gasUsed, gasFee *big.Int) {
//    var (
//        can bool
//        addr crypto.CommonAddress
//        balance, gasPrice *big.Int
//    )
//
//    gasUsed, gasFee = new(big.Int), new(big.Int)
//    subDt := dt.BeginTransaction()
//    can, addr,  _, _, gasPrice = canExecute(subDt, t, nil, CrossChainGas)
//    if !can {
//        return new(big.Int), new(big.Int)
//    }
//
//    cct := &chainTypes.CrossChainTransaction{}
//    err := json.Unmarshal(t.Data.Data, &cct)
//    if err != nil {
//        return new(big.Int), new(big.Int)
//    }
//
//    gasSum := new(big.Int)
//    for _, tx := range cct.Trans {
//        if tx.Data.Type == CrossChainType {
//            continue
//        }
//        g, _ := execute(subDt, tx)
//        gasSum = new(big.Int).Add(gasSum, g)
//    }
//
//    cct.StateRoot = subDt.GetChainStateRoot(database.ChildCHAIN)
//    t.Data.Data, _ = json.Marshal(cct)
//
//    amountSum := new(big.Int).Mul(gasSum, gasPrice)
//    balance = database.GetBalance(addr, t.Data.ChainId)
//    if balance.Cmp(amountSum) >= 0 {
//        gasUsed = new(big.Int).Set(gasSum)
//        gasFee = new(big.Int).Set(amountSum)
//        _, gasFee = deduct(subDt, addr, t.Data.ChainId, balance, gasFee)
//        subDt.Commit()
//    } else {
//        subDt.Discard()
//    }
//
//    return
//}
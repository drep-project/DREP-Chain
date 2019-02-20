package service

import (
    "fmt"
    "github.com/drep-project/drep-chain/common"
    "github.com/drep-project/drep-chain/crypto"
    "github.com/drep-project/drep-chain/log"
    "github.com/drep-project/drep-chain/crypto/secp256k1"
    chainTypes "github.com/drep-project/drep-chain/chain/types"
    chainComponent "github.com/drep-project/drep-chain/chain/component"
    "github.com/drep-project/drep-chain/chain/component/vm"
    "math/big"

    "BlockChainTest/database"
    "encoding/json"
    "bytes"
    "encoding/hex"
    "net/http"
    "strconv"
    "net/url"
)

var (
    childTrans []*chainTypes.Transaction
    lastLeader *secp256k1.PublicKey
    lastMinors []*secp256k1.PublicKey
    lastPrize  *big.Int
)

func (chainService *ChainService) ExecuteTransactions(b *chainTypes.Block) *big.Int {
    if b == nil || b.Header == nil { // || b.Data == nil || b.Data.TxList == nil {
        log.Error("error block nil or header nil")
        return nil
    }
    height := chainService.databaseService.GetMaxHeight()
    if height + 1 != b.Header.Height {
        fmt.Println("error", height, b.Header.Height)
        return nil
    }

    chainService.databaseService.BeginTransaction()
    total := big.NewInt(0)
    if b.Data == nil || b.Data.TxList == nil {
        return total
    }
    for _, t := range b.Data.TxList {
        _, gasFee := chainService.execute(t)
        if t.Data.Type != BlockPrizeType {
            fmt.Println("Delete transaction ", *t)
            fmt.Println(chainService.transactionPool.removeTransaction(t))
        }
        if gasFee != nil {
            total.Add(total, gasFee)
        }
    }

    stateRoot := chainService.databaseService.GetStateRoot();
    if bytes.Equal(b.Header.StateRoot, stateRoot) {
        fmt.Println()
        fmt.Println("matched ", hex.EncodeToString(b.Header.StateRoot), " vs ", hex.EncodeToString(stateRoot))
        height++
        chainService.databaseService.PutMaxHeight(height)
        chainService.databaseService.PutBlock(b)
        chainService.databaseService.Commit()
        fmt.Println("received block: ", true)
        fmt.Println()
        chainService.savePrizeInfo(b, total)
        chainService.preSync(b)
        chainService.doSync(height)
    } else {
        chainService.databaseService.Discard()
        fmt.Println()
        fmt.Println("not matched ", hex.EncodeToString(b.Header.StateRoot), " vs ", hex.EncodeToString(stateRoot))
        fmt.Println("received block: ", false)
        fmt.Println()
    }
    return total
}

func (chainService *ChainService) savePrizeInfo(block *chainTypes.Block, total *big.Int) {
    lastLeader = block.Header.LeaderPubKey
    lastMinors = block.Header.MinorPubKeys
    base := chainService.config.BlockPrize.String()
    basePrize, _ := new(big.Int).SetString(base, 10)
    lastPrize = new(big.Int).Add(basePrize, total)
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
        ChainId:   chainService.chainId,
        StateRoot: chainService.databaseService.GetStateRoot(),
        Trans:     childTrans,
    }
    data, err := json.Marshal(cct)
    if err != nil {
        return
    }
    values := url.Values{}
    values.Add("data", string(data))
    body := values.Encode()
    urlStr := "http://localhost:" + strconv.Itoa(chainService.config.RemotePort) + "/SyncChildChain?" + body
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
    case BlockPrizeType:
        return chainService.executeBlockPrizeTransaction(t)
    //case CrossChainType:
    //   return chainService.executeCrossChainTransaction(t)
    }
    return nil, nil
}

func (chainService *ChainService) canExecute(t *chainTypes.Transaction, gasFloor, gasCap *big.Int) (canExecute bool, addr crypto.CommonAddress, balance, gasLimit, gasPrice *big.Int) {
    addr = crypto.PubKey2Address(t.Data.PubKey)
    balance = chainService.databaseService.GetBalance(addr, t.Data.ChainId, true)
    nonce :=  chainService.databaseService.GetNonce(addr, t.Data.ChainId,true) + 1
    chainService.databaseService.PutNonce(addr, t.Data.ChainId, nonce,true)

    if nonce != t.Data.Nonce {
        return
    }
    if gasFloor != nil {
        amountFloor := new(big.Int).Mul(gasFloor, &t.Data.GasPrice)
        if t.Data.GasLimit.Cmp(gasFloor) < 0 || amountFloor.Cmp(balance) > 0 {
            return
        }
    }
    if gasCap != nil {
        amountCap := new(big.Int).Mul(gasCap, &t.Data.GasPrice)
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
    chainService.databaseService.PutBalance(addr, chainId, leftBalance, true)
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
    if balance.Cmp(&t.Data.Amount) >= 0 {
       to := crypto.Hex2Address(t.Data.To)
       balance = new(big.Int).Sub(balance, &t.Data.Amount)
       balanceTo := chainService.databaseService.GetBalance(to, t.Data.DestChain, true)
       balanceTo = new(big.Int).Add(balanceTo, &t.Data.Amount)
       chainService.databaseService.PutBalance(addr, t.Data.ChainId, balance, true)
       chainService.databaseService.PutBalance(to, t.Data.DestChain, balanceTo, true)
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

    evm := vm.NewEVM(chainService.databaseService)
    returnGas, _ := chainComponent.ApplyTransaction(evm, t)
    gasUsed = new(big.Int).Sub(gasLimit, new(big.Int).SetUint64(returnGas))
    gasFee = new(big.Int).Mul(gasUsed, gasPrice)
    balance = chainService.databaseService.GetBalance(addr, t.Data.ChainId, true)
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

    evm := vm.NewEVM(chainService.databaseService)
    returnGas, _ := chainComponent.ApplyTransaction(evm, t)
    gasUsed = new(big.Int).Sub(gasLimit, new(big.Int).SetUint64(returnGas))
    gasFee = new(big.Int).Mul(gasUsed, gasPrice)
    balance = chainService.databaseService.GetBalance(addr, t.Data.ChainId, true)
    _, gasFee = chainService.deduct(addr, t.Data.ChainId, balance, gasFee)
    return
}

func (chainService *ChainService) executeBlockPrizeTransaction(t *chainTypes.Transaction) (gasUsed *big.Int, gasFee *big.Int) {
    gasUsed, gasFee = new(big.Int), new(big.Int)
    var trans []*chainTypes.Transaction
    if err := json.Unmarshal(t.Data.Data, &trans); err != nil {
        return
    }
    for _, t := range trans {
        addr := crypto.Hex2Address(t.Data.To)
        balance := chainService.databaseService.GetBalance(addr, t.Data.DestChain, true)
        balance = new(big.Int).Add(balance, &t.Data.Amount)
        chainService.databaseService.PutBalance(addr, t.Data.DestChain, balance, true)
    }
    return new(big.Int), new(big.Int)
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
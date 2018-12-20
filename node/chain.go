package node

import (
    "math/big"
    "time"

    "BlockChainTest/log"
    "BlockChainTest/util"
    "BlockChainTest/bean"
    "BlockChainTest/store"
    "BlockChainTest/config"
    "BlockChainTest/network"
    "BlockChainTest/database"
    "BlockChainTest/accounts"
    "encoding/hex"
    "fmt"
)

func SendTransaction(t *bean.Transaction) error {
    peers := store.GetPeers()
    //log.Info("Send transaction")
    if _, offline := network.SendMessage(peers, t); len(offline) == 0 {
        if id, err := t.TxId(); err == nil {
            store.ForwardTransaction(id)
        }
        store.AddTransaction(t)
        store.RemovePeers(offline)
        return nil
    } else {
        return &util.ConnectionError{}
    }
}

//TODO
//发送交易本地nonce, balance 变动

func GenerateBalanceTransaction(to, dc, value string) *bean.Transaction {
    chainId := store.GetChainId()
    destChain := config.Hex2ChainId(dc)
    amount, _ := new(big.Int).SetString(value, 10)
    nonce := database.GetNonce(store.GetAddress(), chainId) + 1
    data := &bean.TransactionData{
        Version: store.Version,
        Nonce: nonce,
        Type: store.TransferType,
        To: to,
        ChainId: chainId,
        DestChain: destChain,
        Amount: amount,
        GasPrice: store.DefaultGasPrice,
        GasLimit: store.TransferGas,
        Timestamp: time.Now().Unix(),
        PubKey: store.GetPubKey(),
    }
    // TODO Get sig bean.transaction{}
    tx := &bean.Transaction{Data: data}
    prvKey := store.GetPrvKey()
    sig, _ := tx.TxSig(prvKey)
    tx.Sig = sig
    return tx
}

func GenerateMinerTransaction(addr, cid string) *bean.Transaction {
    chainId := config.Hex2ChainId(cid)
    nonce := database.GetNonce(store.GetAddress(), chainId) + 1
    data := &bean.TransactionData{
        Nonce:     nonce,
        Type:      store.MinerType,
        ChainId:   chainId,
        GasPrice:  store.DefaultGasPrice,
        GasLimit:  store.MinerGas,
        Timestamp: time.Now().Unix(),
        Data: accounts.Hex2Address(addr).Bytes(),
        PubKey: store.GetPubKey()}
    // TODO Get sig bean.transaction{}
    return &bean.Transaction{Data: data}
}

func GenerateCreateContractTransaction(c string) *bean.Transaction {
    chainId := store.GetChainId()
    nonce := database.GetNonce(store.GetAddress(), chainId) + 1
    code, _ := hex.DecodeString(c)
    data := &bean.TransactionData{
        Nonce: nonce,
        Type: store.CreateContractType,
        ChainId: chainId,
        GasPrice: store.DefaultGasPrice,
        GasLimit: store.CreateContractGas,
        Timestamp: time.Now().Unix(),
        Data: make([]byte, len(code) + 1),
        PubKey: store.GetPubKey(),
    }
    copy(data.Data[1:], code)
    data.Data[0] = 2
    return &bean.Transaction{Data: data}
}


func GenerateCallContractTransaction(addr, cid, in, value string, readOnly bool) *bean.Transaction {
    runningChain := store.GetChainId()
    chainId := config.Hex2ChainId(cid)
    if runningChain != chainId && !readOnly {
        log.Info("you can only call view/pure functions of contract of another chain")
        return &bean.Transaction{}
    }
    nonce := database.GetNonce(store.GetAddress(), runningChain) + 1
    input, _ := hex.DecodeString(in)
    amount, _ := new(big.Int).SetString(value, 10)
    data := &bean.TransactionData{
        Nonce: nonce,
        Type: store.CallContractType,
        ChainId: runningChain,
        DestChain: chainId,
        To: addr,
        Amount: amount,
        GasPrice: store.DefaultGasPrice,
        GasLimit: store.CallContractGas,
        Timestamp: time.Now().Unix(),
        PubKey: store.GetPubKey(),
        Data: make([]byte, len(input) + 1),
    }
    copy(data.Data[1:], input)
    if readOnly {
        data.Data[0] = 1
    } else {
        data.Data[0] = 0
    }
    return &bean.Transaction{Data: data}
}

func GenerateCrossChainTransaction(data []byte) *bean.Transaction {
    transactionData := &bean.TransactionData{
        Version: store.Version,
        Nonce: database.GetNonce(store.GetAddress(), store.GetChainId()) + 1,
        Type: store.CrossChainType,
        ChainId: store.GetChainId(),
        GasPrice: store.DefaultGasPrice,
        GasLimit: store.CrossChainGas,
        Timestamp:time.Now().Unix(),
        Data: data,
        PubKey: store.GetPubKey(),
    }
    return &bean.Transaction{Data: transactionData}
}

func GetH() {
    height := database.GetMaxHeight()
    fmt.Println(height)
}

func GetTxn() {
    h := database.GetMaxHeight()
    var i uint64
    for i = 0; i < h; i++ {
        block := database.GetBlock(i)
        fmt.Println("height: ", i)
        fmt.Println("len: ", len(block.Data.TxList))
        //b, _ := json.Marshal(block)
        //fmt.Println(string(b))
        fmt.Println()
    }
}

//func ForgeCrossChainTransaction() *bean.Transaction {
//    dt := database.BeginTransaction()
//    num := 3
//    trans := make([]*bean.Transaction, num)
//    for i := 0; i < num; i ++ {
//        var data *bean.TransactionData
//        k := rand.Intn(database.CNT)
//        nonce := database.GetNonce(database.ChildADDR[k], database.ChildCHAIN) + 1
//        database.PutNonce(dt, database.ChildADDR[k], database.ChildCHAIN, nonce)
//        data = &bean.TransactionData{
//            Version:   store.Version,
//            Nonce:     nonce,
//            Type:      store.TransferType,
//            To:        database.RootADDR[k].Hex(),
//            ChainId:   database.ChildCHAIN,
//            DestChain: config.RootChain,
//            Amount:    database.AMOUNT[k].Bytes(),
//            GasPrice:  store.DefaultGasPrice.Bytes(),
//            GasLimit:  store.TransferGas.Bytes(),
//            Timestamp: time.Now().Unix(),
//            PubKey:    database.ChildPRV[k].PubKey,
//        }
//        fmt.Println()
//        fmt.Println("transaction ", i, ":")
//        fmt.Println("from:   ", database.ChildADDR[k].Hex(), " ", database.ChildCHAIN.Hex())
//        fmt.Println("to:     ", database.RootADDR[k].Hex(), " ", config.RootChain.Hex())
//        fmt.Println("amount: ", database.AMOUNT[k])
//        fmt.Println()
//        tx := &bean.Transaction{Data: data}
//        prvKey := database.ChildPRV[k]
//        sig, _ := tx.TxSig(prvKey)
//        tx.Sig = sig
//        trans[i] = tx
//    }
//    dt.Discard()
//
//    cct := &bean.CrossChainTransaction{
//        ChainId: database.ChildCHAIN,
//        StateRoot: nil,
//        Trans: trans,
//    }
//    b, _ := json.Marshal(cct)
//
//    ftd := &bean.TransactionData{
//        Version: store.Version,
//        Nonce: database.GetNonce(store.GetAddress(), store.GetChainId()) + 1,
//        Type: store.CrossChainType,
//        ChainId: config.RootChain,
//        GasPrice:store.DefaultGasPrice.Bytes(),
//        GasLimit:store.CrossChainGas.Bytes(),
//        Timestamp:time.Now().Unix(),
//        Data: b,
//        PubKey:store.GetPubKey(),
//    }
//
//    return &bean.Transaction{Data: ftd}
//}
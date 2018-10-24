package node

import (
    "BlockChainTest/bean"
    "BlockChainTest/store"
    "BlockChainTest/network"
    "math/big"
    "time"
    "fmt"
)

func SendTransaction(t *bean.Transaction) error {
    peers := store.GetPeers()
    fmt.Println("Send transaction")
    if err, offline := network.SendMessage(peers, t); err == nil {
        if id, err := t.TxId(); err == nil {
            store.Forward(id)
        }
        store.AddTransaction(t)
        store.RemovePeers(offline)
        return nil
    } else {
        return err
    }
}

func GenerateBalanceTransaction(to bean.Address, amount *big.Int) *bean.Transaction {
    nonce := store.GetNonce(store.GetAddress()) + 1
    data := &bean.TransactionData{
        Nonce:nonce,
        Type:store.TransferType,
        To:string(to),
        Amount:amount.Bytes(),
        GasPrice:store.GasPrice.Bytes(),
        GasLimit:store.TransferGas.Bytes(),
        Timestamp:time.Now().Unix(),
        PubKey:store.GetPubKey()}
    // TODO Get sig bean.Transaction{}
    return &bean.Transaction{Data:data}
}

func GenerateMinerTransaction(addr string) *bean.Transaction {
    nonce := store.GetNonce(store.GetAddress()) + 1
    data := &bean.TransactionData{
        Nonce:     nonce,
        Type:      store.MinerType,
        GasPrice:  store.GasPrice.Bytes(),
        GasLimit:  store.MinerGas.Bytes(),
        Timestamp: time.Now().Unix(),
        Data: []byte(addr),
        PubKey:store.GetPubKey()}
    // TODO Get sig bean.Transaction{}
    return &bean.Transaction{Data: data}
}
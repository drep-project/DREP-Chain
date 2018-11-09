package node

import (
    "BlockChainTest/bean"
    "BlockChainTest/store"
    "BlockChainTest/network"
    "math/big"
    "time"
    "fmt"
    "BlockChainTest/database"
)

func SendTransaction(t *bean.Transaction) error {
    peers := store.GetPeers()
    fmt.Println("Send transaction")
    if err, offline := network.SendMessage(peers, t); err == nil {
        if id, err := t.TxId(); err == nil {
            store.ForwardTransaction(id)
        }
        store.AddTransaction(t)
        store.RemovePeers(offline)
        return nil
    } else {
        return err
    }
}

func GenerateBalanceTransaction(to bean.Address, amount *big.Int) *bean.Transaction {
    nonce := database.GetNonce(bean.Hex2Address(store.GetAddress().String()))
    nonce++
    data := &bean.TransactionData{
        Version: store.Version,
        Nonce:nonce,
        Type:store.TransferType,
        To:to.String(),
        Amount:amount.Bytes(),
        GasPrice:store.GasPrice.Bytes(),
        GasLimit:store.TransferGas.Bytes(),
        Timestamp:time.Now().Unix(),
        PubKey:store.GetPubKey()}
    // TODO Get sig bean.Transaction{}
    tx := &bean.Transaction{Data: data}
    prvKey := store.GetPrvKey()
    sig, _ := tx.TxSig(prvKey)
    tx.Sig = sig
    return tx
}

func GenerateMinerTransaction(addr string) *bean.Transaction {
    nonce := database.GetNonce(bean.Hex2Address(store.GetAddress().String()))
    nonce++
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
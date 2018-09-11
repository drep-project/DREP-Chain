package node

import (
    "BlockChainTest/bean"
    "BlockChainTest/store"
    "BlockChainTest/network"
    "math/big"
    "time"
)

func SendTransaction(t *bean.Transaction)  {
    peers := store.GetPeers()
    network.SendMessage(peers, t)
    store.AddTransaction(t)
}

func GenerateBalanceTransaction(to bean.Address, amount *big.Int) *bean.Transaction {
    nonce := store.GetNonce(store.GetAddress()) + 1
    data := &bean.TransactionData{
        Nonce:nonce, To:string(to),
        Amount:amount.Bytes(),
        GasPrice:store.GasPrice.Bytes(),
        GasLimit:store.TransferGas.Bytes(),
        Timestamp:time.Now().Unix(),
        PubKey:store.GetPubKey()}
    // TODO Get sig bean.Transaction{}
    return &bean.Transaction{Data:data}
}
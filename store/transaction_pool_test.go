package store

import (
    "testing"
    "BlockChainTest/bean"
    "math/big"
    "fmt"
)

func TestPickTransactions(t *testing.T) {
    a0 := bean.Addr(peers[0].PubKey)
    a1 := bean.Addr(peers[1].PubKey)
    balances[a0] = big.NewInt(10000)
    balances[a1] = big.NewInt(20000)
    var tran *bean.Transaction
    tran = &bean.Transaction{Data:&bean.TransactionData{
        To:"0x1", Amount:big.NewInt(1).Bytes(),Nonce:1, GasLimit:TransferGas.Bytes(), GasPrice:GasPrice.Bytes(), PubKey:peers[0].PubKey}}
    AddTransaction(tran)
    tran = &bean.Transaction{Data:&bean.TransactionData{
        To:"0x2", Amount:big.NewInt(3).Bytes(),Nonce:3, GasLimit:TransferGas.Bytes(), GasPrice:GasPrice.Bytes(), PubKey:peers[1].PubKey}}
    AddTransaction(tran)
    tran = &bean.Transaction{Data:&bean.TransactionData{
        To:"0x1", Amount:big.NewInt(3).Bytes(),Nonce:3, GasLimit:TransferGas.Bytes(), GasPrice:GasPrice.Bytes(), PubKey:peers[0].PubKey}}
    AddTransaction(tran)
    tran = &bean.Transaction{Data:&bean.TransactionData{
        To:"0x1", Amount:big.NewInt(2).Bytes(),Nonce:2, GasLimit:TransferGas.Bytes(), GasPrice:GasPrice.Bytes(), PubKey:peers[0].PubKey}}
    AddTransaction(tran)
    tran = &bean.Transaction{Data:&bean.TransactionData{
        To:"0x2", Amount:big.NewInt(2).Bytes(),Nonce:2, GasLimit:TransferGas.Bytes(), GasPrice:GasPrice.Bytes(), PubKey:peers[1].PubKey}}
    AddTransaction(tran)
    print()
}

func print() {
    fmt.Println("Trans:")
    it := trans.Iterator()
    for it.HasNext() {
        e, _ := it.Next().(*bean.Transaction)
        fmt.Println(e)
    }
    fmt.Println("TranSet:")
    for key, value := range tranSet {
        fmt.Println(key, value)
    }
    for key, value := range accountTran {
        fmt.Println(key, value.ToArray())
    }
}
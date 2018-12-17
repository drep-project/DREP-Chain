package node

import (
    "BlockChainTest/store"
    "math/big"
    "errors"
    "BlockChainTest/database"
    "BlockChainTest/accounts"
    "BlockChainTest/config"
    "fmt"
    "encoding/json"
    "BlockChainTest/bean"
)

type ChainApi struct {}

func (chain *ChainApi) Send(addr, destChain, amount string) (string, error) {
    t := GenerateBalanceTransaction(addr, destChain, amount)
    if SendTransaction(t) != nil {
        return "", errors.New("Offline")
    } else {
        return t.TxId()
    }
}

func (chain *ChainApi) CheckBalance(addr accounts.CommonAddress) *big.Int{
    chainId := store.GetChainId()
    return database.GetBalance(addr, chainId)
}

func (chain *ChainApi) CheckNonce(addr accounts.CommonAddress) int64{
    chainId := store.GetChainId()
    return database.GetNonce(addr, chainId)
}

func (chain *ChainApi) Me() *MeInfo{
    addr := store.GetAddress()
    chainId := store.GetChainId()
    nonce := database.GetNonce(addr, chainId)
    balance := database.GetBalance(addr, chainId)
    fmt.Println("check me: ", balance.String())
    fmt.Println("check me: ", nonce)

    //str,_ := json.Marshal(balance)
    //fmt.Println(string(str))
    mi := &MeInfo{
        Address: addr, 
        ChainId: chainId, 
        Nonce: nonce,
        Balance: new(big.Int).Set(balance),
    }
    return mi
}

func (chain *ChainApi) Miner(addr, chainId string) error{
    pk := store.GetPubKey()
    if pk.Equal(store.GetAdminPubKey()) {
        chainId := store.GetChainId().Hex()
        t := GenerateMinerTransaction(addr, chainId)
        if SendTransaction(t) != nil {
            return errors.New("Offline")
        }
    } else {
        return errors.New("You are not allowed.")
    }
    return nil
}

func (chain *ChainApi) Create(code string) (string, error){
    t := GenerateCreateContractTransaction(code)
    if SendTransaction(t) != nil {
        return "", errors.New("Offline")
    } else {
        return t.TxId()
    }
}

func (chain *ChainApi) Call(addr, chainId, input, value string, readOnly bool)  (string, error){
    t := GenerateCallContractTransaction(addr, chainId, input, value, readOnly)
    if SendTransaction(t) != nil {
        return "", errors.New("Offline")
    } else {
        return t.TxId()
    }
}

func (chain *ChainApi) Check(addr accounts.CommonAddress, chainId config.ChainIdType) *accounts.Storage{
    return database.GetStorage(addr, chainId)
}

//func (chain *ChainApi) Cross() (string, error) {
//    t := ForgeCrossChainTransaction()
//    if SendTransaction(t) != nil {
//        return "", errors.New("Offline")
//    } else {
//        return t.TxId()
//    }
//}

func (chain *ChainApi) Travel() {
    itr := database.GetItr()
    for itr.Next() {
        key := itr.Key()
        value := itr.Value()
        storage := &accounts.Storage{}
        json.Unmarshal(value, storage)
        if storage.Balance != nil {
            fmt.Println()
            fmt.Println("key: ", key)
            fmt.Println("value: ", storage)
            fmt.Println()
            continue
        }
        block := &bean.Block{}
        json.Unmarshal(value, block)
        if block.Header == nil || block.Data == nil || block.MultiSig == nil {
            fmt.Println()
            fmt.Println("key: ", key)
            fmt.Println("value: ", value)
            continue
        }
    }
}

type MeInfo struct {
    Address accounts.CommonAddress  `json:"addr"` 
    ChainId config.ChainIdType      `json:"chainId"`
    Nonce int64                     `json:"nonce"` 
    Balance *big.Int                `json:"balance"` 
}
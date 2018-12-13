package node

import (
    "BlockChainTest/store"
    "math/big"
    "encoding/hex"
    "errors"
    "BlockChainTest/database"
    "BlockChainTest/accounts"
)

type ChainApi struct {

}

func (chain *ChainApi) Send(toAddr string, destChain int64, amount int64) (string, error) {
    t := GenerateBalanceTransaction(toAddr, destChain, big.NewInt(amount))
    if SendTransaction(t) != nil {
        return "", errors.New("Offline")
    } else {
        return t.TxId()
    }
}

func (chain *ChainApi) CheckBalance(addr accounts.CommonAddress) *big.Int{
    chainId := store.GetChainId()
    return database.GetBalanceOutsideTransaction(addr, chainId)
}

func (chain *ChainApi) CheckNonce(addr accounts.CommonAddress) int64{
    chainId := store.GetChainId()
    return database.GetNonceOutsideTransaction(addr, chainId)
}

func (chain *ChainApi) Me()  MeInfo{
    addr := store.GetAddress()
    chainId := store.GetChainId()
    nonce := database.GetNonceOutsideTransaction(addr, chainId)
    balance := database.GetBalanceOutsideTransaction(addr, chainId)

    return MeInfo{
        Address: addr, 
        ChainId: chainId, 
        Nonce: nonce,
        Balance: balance, 
    }
}

func (chain *ChainApi) Miner(addr string, chainId int64) error{
    pk := store.GetPubKey()
    if pk.Equal(store.GetAdminPubKey()) {
        chainId := store.GetChainId()
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
    byt, _ := hex.DecodeString(code)
    t := GenerateCreateContractTransaction(byt)

    if SendTransaction(t) != nil {
        return "", errors.New("Offline")
    } else {
        return t.TxId()
    }
}

func (chain *ChainApi) Call(addr accounts.CommonAddress, chainId int64, input string, readOnly bool)  (string, error){
    inp, _ := hex.DecodeString(input)
    t := GenerateCallContractTransaction(addr, chainId, inp, readOnly)
    if SendTransaction(t) != nil {
        return "", errors.New("Offline")
    } else {
        return t.TxId()
    }
}

func (chain *ChainApi) Check(addr accounts.CommonAddress, chainId int64) *accounts.Storage{
    storage := database.GetStorageOutsideTransaction(addr, chainId)
    return storage
}

type MeInfo struct {
    Address accounts.CommonAddress  `json:"addr"` 
    ChainId int64                   `json:"chainId"` 
    Nonce int64                     `json:"nonce"` 
    Balance *big.Int                `json:"balance"` 
}
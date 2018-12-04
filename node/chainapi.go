package node

import (
    "BlockChainTest/store"
    "math/big"
    "errors"
    "BlockChainTest/database"
    "BlockChainTest/accounts"
)

type ChainApi struct {

}

func (chain *ChainApi) Send(toAddr string, destChain int64, amount int64) error {
    chainId := store.GetChainId()
    t := GenerateBalanceTransaction(toAddr, chainId, destChain, big.NewInt(amount))
    if SendTransaction(t) != nil {
        return errors.New("Offline")
    } else {
        return nil
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

func (chain *ChainApi) Me()  MeInfo{
    addr := store.GetAddress()
    chainId := store.GetChainId()
    nonce := database.GetNonce(addr, chainId)
    balance := database.GetBalance(addr, chainId)

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

type MeInfo struct {
    Address accounts.CommonAddress  `json:"addr"` 
    ChainId int64                   `json:"chainId"` 
    Nonce int64                     `json:"nonce"` 
    Balance *big.Int                `json:"balance"` 
}
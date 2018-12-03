package http

import (
    "BlockChainTest/config"
    "BlockChainTest/accounts"
    "BlockChainTest/database"
    "errors"
    "BlockChainTest/store"
    "math/big"
    "BlockChainTest/node"
    "encoding/hex"
    "io/ioutil"
)

func setChain(chainId int64, dataDir string) error {
    return config.SetChain(chainId, dataDir)
}

func createAccount(chainId int64, keystore string) (string, error) {
    runningChain := config.GetChainId()
    existingKeystore := config.GetKeystore()

    IsOnRoot := runningChain == config.RootChain
    IsOnChild := !IsOnRoot
    IsCreatingRoot := chainId == config.RootChain
    IsCreatingChild := !IsCreatingRoot

    if IsOnChild && IsCreatingRoot {
        return "", errors.New("cannot create a child chain account when you are running a root chain process")
    }

    if IsOnChild && IsCreatingChild {
        if chainId != runningChain {
            return "", errors.New("cannot create a child chain account when you are running another child chain")
        }
        _, err := accounts.OpenKeystore(existingKeystore)
        if err == nil {
            return "", errors.New("you have already opened an account under current chain. only one account is permitted under each chain")
        }
        parent, err := accounts.OpenKeystore(keystore)
        if err != nil {
            return "", err
        }
        account, err := accounts.NewNormalAccount(parent, chainId)
        if err != nil {
            return "", err
        }
        err = accounts.SaveKeystore(account.Node, "")
        if err != nil {
            return "", err
        }
        return account.Address.Hex(), nil
    }

    if IsOnRoot && IsCreatingRoot {
        _, err := accounts.OpenKeystore(existingKeystore)
        if err == nil {
            return "", errors.New("you have already opened an account under current chain. only one account is permitted under each chain")
        }
        account, err := accounts.NewNormalAccount(nil, config.RootChain)
        if err != nil {
            return "", err
        }
        err = accounts.SaveKeystore(account.Node, "")
        if err != nil {
            return "", err
        }
        database.PutStorageOutsideTransaction(account.Storage, account.Address, chainId)
        return account.Address.Hex(), nil
    }

    if IsOnRoot && IsCreatingChild {
        parent, err := accounts.OpenKeystore(existingKeystore)
        if err != nil {
            return "", err
        }
        account, err := accounts.NewNormalAccount(parent, chainId)
        if err != nil {
            return "", err
        }
        err = accounts.SaveKeystore(account.Node, keystore)
        if err != nil {
            return "", err
        }
        database.PutStorageOutsideTransaction(account.Storage, account.Address, chainId)
        return account.Address.Hex(), nil
    }

    return "", nil
}

func getAccount() string {
    return accounts.PubKey2Address(store.GetPubKey()).Hex()
}

func getBalance(address string, chainId int64) (*big.Int, error) {
    if config.GetChainId() != config.RootChain {
        return nil, errors.New("you cannot check balance of an account on another child chain while you are running child chain")
    }
    balance := database.GetBalanceOutsideTransaction(accounts.Hex2Address(address), chainId)
    return balance, nil
}

func getNonce(address string, chainId int64) (int64, error) {
    if config.GetChainId() != config.RootChain {
        return -1, errors.New("you cannot check balance of an account on another child chain while you are running child chain")
    }
    nonce := database.GetNonceOutsideTransaction(accounts.Hex2Address(address), chainId)
    return nonce, nil
}

func getMaxHeight() int64 {
    maxHeight := database.GetMaxHeight()
    return maxHeight
}

func sendTransferTransaction(to, amount string, destChain int64) error {
    value, ok := new(big.Int).SetString(amount, 10)
    if !ok {
        return errors.New("params amount parsing error")
    }
    t := node.GenerateBalanceTransaction(to, destChain, value)
    err := node.SendTransaction(t)
    if err != nil {
        return err
    }
    return nil
}

func sendCreateContractTransaction(codeFile string) error {
    code, err := ioutil.ReadFile(codeFile)
    if err != nil {
        return nil
    }
    byteCode, err := hex.DecodeString(string(code))
    if err != nil {
        return err
    }
    t := node.GenerateCreateContractTransaction(byteCode)
    err = node.SendTransaction(t)
    if err != nil {
        return err
    }
    return nil
}
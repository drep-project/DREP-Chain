package node

import (
    "BlockChainTest/bean"
    "BlockChainTest/store"
    "BlockChainTest/network"
    "math/big"
    "time"
    "fmt"
    "BlockChainTest/database"
    "BlockChainTest/accounts"
    "BlockChainTest/config"
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

func GenerateBalanceTransaction(to string, destChain int64, amount *big.Int) *bean.Transaction {
    chainId := config.GetChainId()
    nonce := database.GetNonceOutsideTransaction(accounts.Hex2Address(to), chainId)
    fmt.Println("fefefekfoekfoefkeof", nonce)
    nonce++
    data := &bean.TransactionData{
        Version: store.Version,
        Nonce:nonce,
        Type:store.TransferType,
        To:to,
        ChainId: chainId,
        DestChain: destChain,
        Amount:amount.Bytes(),
        GasPrice:store.GasPrice.Bytes(),
        GasLimit:store.TransferGas.Bytes(),
        Timestamp:time.Now().Unix(),
        PubKey:store.GetPubKey()}
    // TODO Get sig bean.transaction{}
    tx := &bean.Transaction{Data: data}
    prvKey := store.GetPrvKey()
    sig, _ := tx.TxSig(prvKey)
    tx.Sig = sig
    return tx
}

func GenerateMinerTransaction(addr string, chainId int64) *bean.Transaction {
    nonce := database.GetNonceOutsideTransaction(store.GetAddress(), chainId) + 1
    data := &bean.TransactionData{
        Nonce:     nonce,
        Type:      store.MinerType,
        ChainId:   chainId,
        GasPrice:  store.GasPrice.Bytes(),
        GasLimit:  store.MinerGas.Bytes(),
        Timestamp: time.Now().Unix(),
        Data: accounts.Hex2Address(addr).Bytes(),
        PubKey:store.GetPubKey()}
    // TODO Get sig bean.transaction{}
    return &bean.Transaction{Data: data}
}

func GenerateCreateContractTransaction(code []byte) *bean.Transaction {
    chainId := config.GetChainId()
    nonce := database.GetNonceOutsideTransaction(store.GetAddress(), chainId) + 1
    data := &bean.TransactionData{
        Nonce: nonce,
        Type: store.CreateContractType,
        ChainId: chainId,
        GasPrice: store.GasPrice.Bytes(),
        GasLimit: store.CreateContractGas.Bytes(),
        Timestamp: time.Now().Unix(),
        Data: make([]byte, len(code) + 1),
        PubKey: store.GetPubKey(),
    }
    copy(data.Data[1:], code)
    data.Data[0] = 2
    return &bean.Transaction{Data: data}
}

func GenerateCallContractTransaction(addr accounts.CommonAddress, chainId int64, input []byte, readOnly bool) *bean.Transaction {
    runningChain := config.GetChainId()
    nonce := database.GetNonceOutsideTransaction(store.GetAddress(), runningChain) + 1
    if runningChain != chainId && !readOnly {
        fmt.Println("you can only call view/pure functions of contract of another chain")
        return &bean.Transaction{}
    }
    data := &bean.TransactionData{
        Nonce: nonce,
        Type: store.CallContractType,
        ChainId: runningChain,
        DestChain: chainId,
        To: addr.Hex(),
        GasPrice: store.GasPrice.Bytes(),
        GasLimit: store.CallContractGas.Bytes(),
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
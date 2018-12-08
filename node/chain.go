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
    "BlockChainTest/mycrypto"
    "math/rand"
    "fmt"
)

const (
    cnt = 50
)

var (
    rp [cnt]*mycrypto.PrivateKey
    ra [cnt]accounts.CommonAddress
    cp [cnt]*mycrypto.PrivateKey
    ca [cnt]accounts.CommonAddress
    cc [cnt]int64
    amount [cnt]*big.Int
)

//func init() {
//    for i := 0; i < cnt; i++ {
//        rp[i], _ = mycrypto.GeneratePrivateKey()
//        ra[i] = accounts.PubKey2Address(rp[i].PubKey)
//        cp[i], _ = mycrypto.GeneratePrivateKey()
//        ca[i] = accounts.PubKey2Address(cp[i].PubKey)
//        cc[i] = rand.Int63n(1000) + 123
//        //cc[i] = 0
//        amount[i] = new(big.Int).SetInt64(10000 + int64(i) * 100)
//        database.PutBalanceOutSideTransaction(ra[i], accounts.RootChainID, new(big.Int).SetInt64(100000000))
//        database.PutBalanceOutSideTransaction(ca[i], cc[i], new(big.Int).SetInt64(100000000))
//    }
//}

func SendTransaction(t *bean.Transaction) error {
    peers := store.GetPeers()
    log.Info("Send transaction")
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

func GenerateBalanceTransaction(to string, destChain int64, amount *big.Int) *bean.Transaction {
    chainId := config.GetChainId()
    nonce := database.GetNonceOutsideTransaction(store.GetAddress(), chainId) + 1
    data := &bean.TransactionData{
        Version: store.Version,
        Nonce:nonce,
        Type:store.TransferType,
        To:to,
        ChainId: chainId,
        DestChain: destChain,
        Amount:amount.Bytes(),
        GasPrice:store.DefaultGasPrice.Bytes(),
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
        GasPrice:  store.DefaultGasPrice.Bytes(),
        GasLimit:  store.MinerGas.Bytes(),
        Timestamp: time.Now().Unix(),
        Data: accounts.Hex2Address(addr).Bytes(),
        PubKey:store.GetPubKey()}
    // TODO Get sig bean.transaction{}
    return &bean.Transaction{Data: data}
}

func GenerateCreateContractTransaction(code []byte) *bean.Transaction {
    chainId := config.GetConfig().ChainId
    nonce := database.GetNonceOutsideTransaction(store.GetAddress(), chainId) + 1
    data := &bean.TransactionData{
        Nonce: nonce,
        Type: store.CreateContractType,
        ChainId: chainId,
        GasPrice: store.DefaultGasPrice.Bytes(),
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
    runningChain := config.GetConfig().ChainId
    nonce := database.GetNonceOutsideTransaction(store.GetAddress(), runningChain) + 1
    if runningChain != chainId && !readOnly {
        log.Info("you can only call view/pure functions of contract of another chain")
        return &bean.Transaction{}
    }
    data := &bean.TransactionData{
        Nonce: nonce,
        Type: store.CallContractType,
        ChainId: runningChain,
        DestChain: chainId,
        To: addr.Hex(),
        GasPrice: store.DefaultGasPrice.Bytes(),
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

func ForgeTransferTransaction() []*bean.Transaction {
    dbTran := database.BeginTransaction()
    num := 100
    trans := make([]*bean.Transaction, num)
    for i := 0; i < num; i ++ {
        transferDirection := rand.Intn(2)
        k := rand.Intn(cnt)
        var data *bean.TransactionData
        if transferDirection == 1 {
            nonce := database.GetNonceInsideTransaction(dbTran, ra[k], accounts.RootChainID) + 1
            database.PutNonceInsideTransaction(dbTran, ra[k], accounts.RootChainID, nonce)
            data = &bean.TransactionData{
                Version:   store.Version,
                Nonce:     nonce,
                Type:      store.TransferType,
                To:        ca[k].Hex(),
                ChainId:   accounts.RootChainID,
                DestChain: cc[k],
                Amount:    amount[k].Bytes(),
                GasPrice:  store.DefaultGasPrice.Bytes(),
                GasLimit:  store.TransferGas.Bytes(),
                Timestamp: time.Now().Unix(),
                PubKey:    rp[k].PubKey,
            }
            fmt.Println()
            fmt.Println("transaction ", i, ":")
            fmt.Println("from:   ", ra[k].Hex(), " ", accounts.RootChainID)
            fmt.Println("to:     ", ca[k].Hex(), " ", cc[k])
            fmt.Println("amount: ", amount[k])
            fmt.Println()
        } else {
            nonce := database.GetNonceInsideTransaction(dbTran, ca[k], cc[k]) + 1
            database.PutNonceInsideTransaction(dbTran, ra[k], accounts.RootChainID, nonce)
            data = &bean.TransactionData{
                Version:   store.Version,
                Nonce:     nonce,
                Type:      store.TransferType,
                To:        ra[k].Hex(),
                ChainId:   cc[k],
                DestChain: accounts.RootChainID,
                Amount:    amount[k].Bytes(),
                GasPrice:  store.DefaultGasPrice.Bytes(),
                GasLimit:  store.TransferGas.Bytes(),
                Timestamp: time.Now().Unix(),
                PubKey:    cp[k].PubKey,
            }
            fmt.Println()
            fmt.Println("transaction ", i, ":")
            fmt.Println("from:   ", ca[k].Hex(), " ", cc[k])
            fmt.Println("to:     ", ra[k].Hex(), " ", accounts.RootChainID)
            fmt.Println("amount: ", amount[k])
            fmt.Println()
        }
        tx := &bean.Transaction{Data: data}
        prvKey := store.GetPrvKey()
        sig, _ := tx.TxSig(prvKey)
        tx.Sig = sig
        trans[i] = tx
    }
    dbTran.Discard()
    return trans
}

//TODO
//删除trie上的非account信息

package store

import (
    "testing"
    "BlockChainTest/database"
    "math/big"
    "log"
    "BlockChainTest/accounts"
    "BlockChainTest/bean"
    "BlockChainTest/mycrypto"
    "math/rand"
    "time"
    "fmt"
)

var (
    accountNum    = 1000
    txNum         = 10000
    cid     int64 = 0
    bal           = "1000000000000000000"
    transactions  []*bean.Transaction
    pubKeys       []*mycrypto.Point
    addrs         []accounts.CommonAddress
)

func forgeU() {
    pubKeys = make([]*mycrypto.Point, accountNum)
    addrs = make([]accounts.CommonAddress, accountNum)
    dt := database.BeginTransaction()
    for i := 0; i < accountNum; i++ {
        prvKey, _ := mycrypto.GeneratePrivateKey()
        pubKeys[i] = prvKey.PubKey
        addrs[i] = accounts.PubKey2Address(prvKey.PubKey)
        balance, _ := new(big.Int).SetString(bal, 10)
        database.PutBalanceInsideTransaction(dt, addrs[i], chainId, balance)
    }
    dt.Commit()
    //itr := database.GetItr()
    //for itr.Next() {
    //    key := itr.Key()
    //    value := itr.Value()
    //    fmt.Println(hex.EncodeToString(key), ": ", string(value))
    //}
}

func forgeT() {
    dt := database.BeginTransaction()
    transactions = make([]*bean.Transaction, txNum)
    for i := 0; i < txNum; i++ {
        ff := rand.Intn(accountNum)
        tt := rand.Intn(accountNum)
        if tt == ff {
            if ff > 0 {
                tt = ff - 1
            } else {
                tt = ff + 1
            }
        }
        nonce := database.GetNonceInsideTransaction(dt, addrs[ff], cid) + 1
        database.PutNonceInsideTransaction(dt, addrs[ff], cid, nonce)
        data := &bean.TransactionData{
            Version: Version,
            Nonce: nonce,
            Type: TransferType,
            To: addrs[tt].Hex(),
            ChainId: cid,
            DestChain: cid,
            Amount: new(big.Int).SetInt64(rand.Int63n(100) + 1).Bytes(),
            GasPrice: DefaultGasPrice.Bytes(),
            GasLimit: TransferGas.Bytes(),
            Timestamp: time.Now().Unix(),
            PubKey: pubKeys[ff],
        }
        transactions[i] = &bean.Transaction{Data: data}
    }
    dt.Discard()
    //for i, t := range transactions {
    //    b, _ := json.Marshal(t)
    //    fmt.Println(i + 1, string(b))
    //}
    //itr := database.GetItr()
    //for itr.Next() {
    //   key := itr.Key()
    //   value := itr.Value()
    //   fmt.Println(hex.EncodeToString(key), ": ", string(value))
    //}
}

func executeT(t *bean.Transaction, dt *database.Transaction) *big.Int {
    addr := accounts.PubKey2Address(t.Data.PubKey)
    nonce := t.Data.Nonce
    curN := database.GetNonceInsideTransaction(dt, addr, t.Data.ChainId)
    if curN + 1 != nonce {
        fmt.Println("return 1")
        return new(big.Int)
    }
    database.PutNonceInsideTransaction(dt, addr, t.Data.ChainId, curN + 1)
    gasPrice := big.NewInt(0).SetBytes(t.Data.GasPrice)
    gasLimit := big.NewInt(0).SetBytes(t.Data.GasLimit)
    gasFee := big.NewInt(0).Mul(gasLimit, gasPrice)
    balance := database.GetBalanceInsideTransaction(dt, addr, t.Data.ChainId)
    if gasFee.Cmp(balance) > 0 {
        log.Fatal("Error, gas not right")
        return new(big.Int)
    }
    if gasLimit.Cmp(TransferGas) < 0 {
        balance.Sub(balance, gasFee)
        return gasLimit
    } else {
        amount := big.NewInt(0).SetBytes(t.Data.Amount)
        total := big.NewInt(0).Add(amount, gasFee)
        if balance.Cmp(total) >= 0 {
            balance.Sub(balance, total)
            to := t.Data.To
            balance2 := database.GetBalanceInsideTransaction(dt, accounts.Hex2Address(to), t.Data.ChainId)
            balance2.Add(balance2, amount)
            database.PutBalanceInsideTransaction(dt, accounts.Hex2Address(to), t.Data.ChainId, balance2)
        } else {
            balance.Sub(balance, gasFee)
        }
    }
    database.PutBalanceInsideTransaction(dt, addr, t.Data.ChainId, balance)
    return TransferGas
}

func TestMountOfTransactions(t *testing.T) {
    t1 := time.Now()
    forgeU()
    t2 := time.Now()
    fmt.Println("gap t1 t2: ", t2.Sub(t1))

    t3 := time.Now()
    forgeT()
    t4 := time.Now()
    fmt.Println("gap t3 t4: ", t4.Sub(t3))

    t5 := time.Now()
    dt := database.BeginTransaction()
    for _, t := range transactions {
        executeT(t, dt)
        //gasFee := executeT(t, dt)
        //fmt.Println("gas fee: ", gasFee)
    }
    dt.Commit()
    t6 := time.Now()
    fmt.Println("gap t5 t6: ", t6.Sub(t5))
}

package browser

import (
    "testing"
    "BlockChainTest/bean"
    "math/rand"
    "math/big"
    "time"
    "BlockChainTest/mycrypto"
    "fmt"
)

var (
    n = 100
    blocks = make([]*bean.Block, n)
    transactions = make([]*bean.Transaction, n)
)

func GenPubKey() *mycrypto.Point {
    cur := mycrypto.GetCurve()
    prv := new(big.Int).SetInt64(rand.Int63()).Bytes()
    pubKey := cur.ScalarBaseMultiply(prv)
    return pubKey
}

func GenTransactions() {
    for i:=0; i<n; i++ {
        data := &bean.TransactionData{
            Version: 1,
            Nonce: rand.Int63(),
            Type: 1,
            To: bean.Bytes2Address(mycrypto.Hash256(new(big.Int).SetInt64(rand.Int63()).Bytes())).Hex(),
            ChainId: rand.Int63(),
            Amount: new(big.Int).SetInt64(rand.Int63()).Bytes(),
            GasPrice: new(big.Int).SetInt64(rand.Int63()).Bytes(),
            GasLimit: new(big.Int).SetInt64(rand.Int63()).Bytes(),
            Timestamp: time.Now().Unix() + int64(rand.Int31()),
            Data: make([]byte, 0),
            PubKey: GenPubKey(),
        }
        transactions[i] = &bean.Transaction{Data: data}
    }
}

func GenTxHashes() [][]byte {
    m := rand.Intn(n - 1) + 1
    ret := make([][]byte, m)
    for i:=0; i<m; i++ {
        k := rand.Intn(n)
        tx := transactions[k]
        ret[i], _ = tx.TxHash()
    }
    return ret
}

func GenBlocks() {
    for i:=0; i<n; i++ {
        header := &bean.BlockHeader{
            ChainId: rand.Int63(),
            Version: 1,
            PreviousHash: mycrypto.Hash256(new(big.Int).SetInt64(rand.Int63()).Bytes()),
            GasLimit: new(big.Int).SetInt64(rand.Int63()).Bytes(),
            GasUsed: new(big.Int).SetInt64(rand.Int63()).Bytes(),
            Height: rand.Int63(),
            Timestamp: time.Now().Unix() + int64(rand.Int31()),
            StateRoot: make([]byte, 0),
            MerkleRoot: make([]byte, 0),
            TxHashes: GenTxHashes(),
            LeaderPubKey: nil,
            MinorPubKeys: nil,
        }
        blocks[i] = &bean.Block{Header: header}
    }
}

func RandomSelectTxWeb() string {
    m := rand.Intn(n - 1) + 1
    ret := make([]string, m)
    for i:=0; i<m; i++ {
        k := rand.Intn(n)
        ret[i] = ParseTransaction(transactions[k])
    }
    return ret[0]
}

func RandomSelectBlockWeb() string {
    m := rand.Intn(n - 1) + 1
    ret := make([]string, m)
    for i:=0; i<m; i++ {
        k := rand.Intn(n)
        ret[i] = ParseBlock(blocks[k])
    }
    return ret[0]
}

func TestTx(t *testing.T) {
    GenTransactions()
    GenBlocks()
    ret := RandomSelectTxWeb()
    fmt.Println(ret)
}

func TestBlock(t *testing.T) {
    GenTransactions()
    GenBlocks()
    ret := RandomSelectBlockWeb()
    fmt.Println(ret)
}

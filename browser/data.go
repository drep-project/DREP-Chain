package browser

import (
    "BlockChainTest/bean"
    "encoding/hex"
    "math/big"
    "encoding/json"
    "math/rand"
)

type BlockWeb struct {
    ChainId      int64
    Height       int64
    Timestamp    int64
    Hash         string
    PreviousHash string
    GasUsed      string
    GasLimit     string
    TxHashes     []string
    Size         int
}

func ParseBlock(block *bean.Block) string {
    b := &BlockWeb{}
    b.ChainId = block.Header.ChainId
    b.Height = block.Header.Height
    b.Timestamp = block.Header.Timestamp
    b.Hash, _ = block.BlockHash()
    b.PreviousHash = "0x" + hex.EncodeToString(block.Header.PreviousHash)
    b.GasUsed = new(big.Int).SetBytes(block.Header.GasUsed).String()
    b.GasLimit = new(big.Int).SetBytes(block.Header.GasLimit).String()
    b.TxHashes = block.TxHashes()
    j, _ := json.Marshal(b)
    b.Size = len(j)
    ret, _ := json.Marshal(b)
    return string(ret)
}

type TransactionWeb struct {
    Nonce     int64
    Timestamp int64
    Hash      string
    From      string
    To        string
    ChainId   int64
    Amount    string
    GasPrice  string
    GasUsed   string
    GasLimit  string
    Height    int64
}

func ParseTransaction(tx *bean.Transaction, height int64) string {
    t := &TransactionWeb{}
    t.Nonce = tx.Data.Nonce
    t.Timestamp = tx.Data.Timestamp
    h, _ := tx.TxHash()
    t.Hash = "0x" + hex.EncodeToString(h)
    t.From = "0x" + bean.PubKey2Address(tx.Data.PubKey).Hex()
    t.To = "0x" + tx.Data.To
    t.ChainId = tx.Data.ChainId
    t.Amount = new(big.Int).SetBytes(tx.Data.Amount).String()
    t.GasPrice = new(big.Int).SetBytes(tx.Data.GasPrice).String()
    t.GasUsed = new(big.Int).SetInt64(rand.Int63()).String()
    t.GasLimit = new(big.Int).SetBytes(tx.Data.GasLimit).String()
    t.Height = height
    ret, _ := json.Marshal(t)
    return string(ret)
}
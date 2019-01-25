package http

import (
    "BlockChainTest/bean"
    "encoding/hex"
    "math/big"
    "math/rand"
    "strconv"
    "BlockChainTest/accounts"
)

type BlockWeb struct {
    ChainId      string
    Height       int64
    Timestamp    int64
    Hash         string
    PreviousHash string
    GasUsed      string
    GasLimit     string
    Size         int
    MiningLeader string
    MiningMember []string
}

func ParseBlock(block *bean.Block) *BlockWeb {
    b := &BlockWeb{}
    b.ChainId = "0x" + strconv.FormatInt(block.Header.ChainId,10)
    b.Height = block.Header.Height
    b.Timestamp = block.Header.Timestamp * 1000
    b.Hash, _ = block.BlockHash()
    b.PreviousHash = "0x" + hex.EncodeToString(block.Header.PreviousHash)
    b.GasUsed = new(big.Int).SetBytes(block.Header.GasUsed).String()
    b.GasLimit = new(big.Int).SetBytes(block.Header.GasLimit).String()
    b.Size = int(block.Data.TxCount)
    var minorPubKeys []string

    var leaderPubKey = "0x" + accounts.PubKey2Address(block.Header.LeaderPubKey).Hex()
    for _, key := range(block.Header.MinorPubKeys) {
        minorPubKeys = append(minorPubKeys, "0x" + accounts.PubKey2Address(key).Hex())
    }
    b.MiningMember = minorPubKeys
    b.MiningLeader = leaderPubKey
    return b
}

type TransactionWeb struct {
    Nonce     int64
    Timestamp int64
    Hash      string
    From      string
    To        string
    ChainId   string
    Amount    string
    GasPrice  string
    GasUsed   string
    GasLimit  string
    Data      string
}

func ParseTransaction(tx *bean.Transaction) *TransactionWeb {
    t := &TransactionWeb{}
    t.Nonce = tx.Data.Nonce
    t.Timestamp = tx.Data.Timestamp * 1000
    h, _ := tx.TxHash()
    t.Hash = "0x" + hex.EncodeToString(h)
    if tx.Data.PubKey != nil {
        t.From = "0x" + accounts.PubKey2Address(tx.Data.PubKey).Hex()
    }
    t.To = "0x" + tx.Data.To
    t.Data = "0x" + hex.EncodeToString(tx.Data.Data)
    t.ChainId = "0x" + strconv.FormatInt(tx.Data.ChainId,10)
    t.Amount = new(big.Int).SetBytes(tx.Data.Amount).String()
    t.GasPrice = new(big.Int).SetBytes(tx.Data.GasPrice).String()
    t.GasUsed = new(big.Int).SetInt64(rand.Int63()).String()
    t.GasLimit = new(big.Int).SetBytes(tx.Data.GasLimit).String()
    return t
}
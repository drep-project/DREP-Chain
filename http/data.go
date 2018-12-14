package rest

import (
    "BlockChainTest/bean"
    "encoding/hex"
    "math/big"
    "math/rand"
    "BlockChainTest/config"
    "BlockChainTest/accounts"
)

type BlockWeb struct {
    ChainId      config.ChainIdType
    Height       int64
    Timestamp    int64
    Hash         string
    PreviousHash string
    GasUsed      string
    GasLimit     string
    TxHashes     []string
    Size         int
    MiningLeader string
    MiningMember []string
}

func ParseBlock(block *bean.Block) *BlockWeb {
    b := &BlockWeb{}
    b.ChainId = block.Header.ChainId
    b.Height = block.Header.Height
    b.Timestamp = block.Header.Timestamp * 1000
    b.Hash, _ = block.BlockHash()
    b.PreviousHash = "0x" + hex.EncodeToString(block.Header.PreviousHash)
    b.GasUsed = new(big.Int).SetBytes(block.Header.GasUsed).String()
    b.GasLimit = new(big.Int).SetBytes(block.Header.GasLimit).String()

    var minorPubKeys []string

    var leaderPubKey = "0x" + accounts.PubKey2Address(block.Header.LeaderPubKey).Hex()
    for _, key := range(block.Header.MinorPubKeys) {
        minorPubKeys = append(minorPubKeys, "0x" + accounts.PubKey2Address(key).Hex())
    }
    b.MiningMember = minorPubKeys
    b.MiningLeader = leaderPubKey

    b.TxHashes = block.TxHashes()
    return b
}

type TransactionWeb struct {
    Nonce     int64
    Timestamp int64
    Hash      string
    From      string
    To        string
    ChainId   config.ChainIdType
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
<<<<<<< HEAD:http/data.go
    t.From = "0x" + bean.PubKey2Address(tx.Data.PubKey).Hex()
    t.To = tx.Data.To
=======
    t.From = "0x" + accounts.PubKey2Address(tx.Data.PubKey).Hex()
    t.To = "0x" + tx.Data.To
>>>>>>> 39bb07a... modify chainId type and add revert:network/http/data.go
    t.Data = "0x" + hex.EncodeToString(tx.Data.Data)
    t.ChainId = tx.Data.ChainId
    t.Amount = new(big.Int).SetBytes(tx.Data.Amount).String()
    t.GasPrice = new(big.Int).SetBytes(tx.Data.GasPrice).String()
    t.GasUsed = new(big.Int).SetInt64(rand.Int63()).String()
    t.GasLimit = new(big.Int).SetBytes(tx.Data.GasLimit).String()
    return t
}
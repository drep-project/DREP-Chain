package browser

import (
    "BlockChainTest/bean"
    "encoding/hex"
    "math/big"
    "encoding/json"
)

type BlockWeb struct {
    height int64
    timestamp int64
    hash string
    previousHash string
    gasUsed string
    gasLimit string
    txHashes []string
    size int
}

func ParseBlock(block *bean.Block) string {
    b := &BlockWeb{}
    b.height = block.Header.Height
    b.timestamp = block.Header.Timestamp
    b.hash, _ = block.BlockHash()
    b.previousHash = "0x" + hex.EncodeToString(block.Header.PreviousHash)
    b.gasUsed = new(big.Int).SetBytes(block.Header.GasUsed).String()
    b.gasLimit = new(big.Int).SetBytes(block.Header.GasLimit).String()
    b.txHashes = block.TxHashes()
    j, _ := json.Marshal(b)
    b.size = len(j)
    ret, _ := json.Marshal(b)
    return string(ret)
}

type TransactionWeb struct {
    nonce int64
    timestamp int64
    hash string
    from string
    to string
    amount string
    gasPrice string
    gasUsed string
    gasLimit string
}

func ParseTransaction(tx *bean.Transaction) string {
    t := &TransactionWeb{}
    t.nonce = tx.Data.Nonce
    t.timestamp = tx.Data.Timestamp
    h, _ := tx.TxHash()
    t.hash = "0x" + hex.EncodeToString(h)
    t.from = "0x" + bean.PubKey2Address(tx.Data.PubKey).Hex()
    t.to = "0x" + tx.Data.To
    t.amount = new(big.Int).SetBytes(tx.Data.Amount).String()
    t.gasPrice = new(big.Int).SetBytes(tx.Data.GasPrice).String()
    t.gasUsed = tx.GetGasUsed().String()
    t.gasLimit = new(big.Int).SetBytes(tx.Data.GasLimit).String()
    ret, _ := json.Marshal(t)
    return string(ret)
}

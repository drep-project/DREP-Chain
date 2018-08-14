package common

import "math/big"

const (
    MSG_BLOCK = 1
    MSG_TRANSACTION = 2
)
type Message struct {
    Type int
    Body interface{}
}

type BlockHeader struct {
    Version      int32
    PreviousHash []byte
    GasLimit     big.Int
    GasUsed      big.Int
    Height       int32
    CreatedTime  int32
    MerkleRoot   []byte
    TxHashes     [][]byte
    LeaderPubKey []byte
    MinorPubKeys [][]byte
}

type BlockData struct {
    TxCount int32
    TxList  []*Transaction
}

type Block struct {
    Header   *BlockHeader
    Data     *BlockData
    MultiSig []byte
}

type Transaction struct {
    Version      int32
    Nonce        int64
    ToAddress    []byte
    Amount       big.Int
    GasPrice     big.Int
    GasLimit     big.Int
    ProducedTime int32
    PubKey       []byte
    Sig          []byte
}

func (t *Transaction) GetId() string {
    return ""
}

type BlockChain struct {
    Blocks []Block
}


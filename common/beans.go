package common

import "math/big"

const (
    MSG_SETUP1 = 1 // miner send, MS
    MSG_BLOCK1_COMMIT = 3 // MS
    MSG_BLOCK1_CHALLENGE = 4 // LS,
    MSG_BLOCK1_RESPONSE = 5 // MS

    MSG_SETUP2 = 6
    MSG_BLOCK2_COMMIT = 7 // MS
    MSG_BLOCK2_CHALLENGE = 8 // LS,
    MSG_BLOCK2_RESPONSE = 9 // MS

    MSG_BLOCK = 9
    MSG_TRANSACTION = 10
)
type Message struct {
    Type int
    Body interface{}
}

type Address string

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



package bean

import (
    "encoding/hex"
    "BlockChainTest/mycrypto"
    "math/big"
    "encoding/json"
)

type BlockHeader struct {
    Version              int32
    PreviousHash         []byte
    GasLimit             []byte
    GasUsed              []byte
    Height               int64
    Timestamp            int64
    StateRoot            []byte
    MerkleRoot           []byte
    TxHashes             [][]byte
    LeaderPubKey         *mycrypto.Point
    MinorPubKeys         []*mycrypto.Point
}

type BlockData struct {
    TxCount              int32
    TxList               []*Transaction
}

type Block struct {
    Header               *BlockHeader
    Data                 *BlockData
    MultiSig             *mycrypto.Signature
}

type TransactionData struct {
    Version              int32
    Nonce                int64
    Type                 int32
    To                   string
    Amount               []byte
    GasPrice             []byte
    GasLimit             []byte
    Timestamp            int64
    Data                 []byte
    PubKey               *mycrypto.Point
}

type Transaction struct {
    Data                 *TransactionData
    Sig                  *mycrypto.Signature
}

type MultiSignature struct {
    Sig                  *mycrypto.Signature
    Bitmap               []byte
}

func (tx *Transaction) TxId() (string, error) {
    b, err := json.Marshal(tx.Data)
    if err != nil {
        return "", err
    }
    id := hex.EncodeToString(mycrypto.Hash256(b))
    return id, nil
}

func (tx *Transaction) TxHash() ([]byte, error) {
    b, err := json.Marshal(tx)
    if err != nil {
        return nil, err
    }
    h := mycrypto.Hash256(b)
    return h, nil
}

func (tx *Transaction) TxSig(prvKey *mycrypto.PrivateKey) (*mycrypto.Signature, error) {
    b, err := json.Marshal(tx.Data)
    if err != nil {
        return nil, err
    }
    return mycrypto.Sign(prvKey, b)
}

func (tx *Transaction) Address() CommonAddress {
    return PubKey2Address(tx.Data.PubKey)
}

func (tx *Transaction) Addr() Address {
    return Addr(tx.Data.PubKey)
}

func (tx *Transaction) GetGasQuantity() *big.Int {
    return new(big.Int).SetInt64(int64(100))
}

func (tx *Transaction) GetGasUsed() *big.Int {
    gasQuantity := tx.GetGasQuantity()
    gasPrice := new(big.Int).SetBytes(tx.Data.GasPrice)
    gasUsed := new(big.Int).Mul(gasQuantity, gasPrice)
    return gasUsed
}

func (block *Block) BlockID() (string, error) {
    b, err := json.Marshal(block.Header)
    if err != nil {
        return "", err
    }
    id := hex.EncodeToString(mycrypto.Hash256(b))
    return id, nil
}

func Height2Key(height int64) string {
    return hex.EncodeToString(new(big.Int).SetInt64(height).Bytes())
}

func MarshalBlock(block *Block) ([]byte, error) {
    b, err := json.Marshal(block)
    if err != nil {
        return nil, err
    }
    return b, nil
}

func UnmarshalBlock(b []byte) (*Block, error) {
    block := &Block{}
    err := json.Unmarshal(b, block)
    if err != nil {
        return nil, err
    }
    return block, nil
}
package types

import (
    "github.com/drep-project/drep-chain/common"
    "github.com/drep-project/drep-chain/crypto/secp256k1"
    "github.com/drep-project/drep-chain/crypto"
    "github.com/drep-project/drep-chain/crypto/sha3"
    "encoding/hex"
    "encoding/json"
    "math/big"
)

type BlockHeader struct {
    ChainId              common.ChainIdType
    Version              int32
    PreviousHash         []byte
    GasLimit             []byte
    GasUsed              []byte
    Height               int64
    Timestamp            int64
    StateRoot            []byte
    MerkleRoot           []byte
    TxHashes             [][]byte
    LeaderPubKey         *secp256k1.PublicKey
    MinorPubKeys         []*secp256k1.PublicKey
}

type BlockData struct {
    TxCount              int32
    TxList               []*Transaction
}

type Block struct {
    Header               *BlockHeader
    Data                 *BlockData
    MultiSig             *MultiSignature
}

type TransactionData struct {
    Version              int32
    Nonce                int64
    Type                 int32
    To                   string
    ChainId              common.ChainIdType
    DestChain            common.ChainIdType
    Amount               big.Int
    GasPrice             big.Int
    GasLimit             big.Int
    Timestamp            int64
    Data                 []byte
    PubKey               *secp256k1.PublicKey
}

type Transaction struct {
    Data                 *TransactionData
    Sig                  []byte
}

type CrossChainTransaction struct {
    ChainId   common.ChainIdType
    StateRoot []byte
    Trans     []*Transaction
}

type Log struct {
    Address crypto.CommonAddress
    ChainId common.ChainIdType
    TxHash  []byte
    Topics  [][]byte
    Data    []byte
}

type MultiSignature struct {
    Sig                  secp256k1.Signature
    Bitmap               []byte
}

func (tx *Transaction) TxId() (string, error) {
    b, err := json.Marshal(tx.Data)
    if err != nil {
        return "", err
    }
    id := hex.EncodeToString(sha3.Hash256(b))
    return id, nil
}

func (tx *Transaction) TxHash() ([]byte, error) {
    b, err := json.Marshal(tx)
    if err != nil {
        return nil, err
    }
    h := sha3.Hash256(b)
    return h, nil
}

func (tx *Transaction) TxSig(prvKey *secp256k1.PrivateKey) (*secp256k1.Signature, error) {
    b, err := json.Marshal(tx.Data)
    if err != nil {
        return nil, err
    }

    return prvKey.Sign(sha3.Hash256(b))
}

func (tx *Transaction) GetGasUsed() *big.Int {
    return new(big.Int).SetInt64(int64(100))
}

func (tx *Transaction) GetGas() *big.Int {
    gasQuantity := tx.GetGasUsed()
    gasUsed := new(big.Int).Mul(gasQuantity, &tx.Data.GasPrice)
    return gasUsed
}

func (block *Block) BlockHash() ([]byte, error) {
    b, err := json.Marshal(block.Header)
    if err != nil {
        return nil, err
    }
    return sha3.Hash256(b), nil
}

func (block *Block) BlockHashHex() (string, error) {
    h, err := block.BlockHash()
    if err != nil {
        return "", err
    }
    return "0x" + hex.EncodeToString(h), nil
}

func (block *Block) TxHashes() []string {
    th := make([]string, len(block.Header.TxHashes))
    for i, hash := range block.Header.TxHashes {
        th[i] = "0x" + hex.EncodeToString(hash)
    }
    return th
}
package types

import (
    "encoding/hex"
    "encoding/json"
    "github.com/drep-project/drep-chain/app"
    "github.com/drep-project/drep-chain/crypto/secp256k1"
    "github.com/drep-project/drep-chain/crypto/sha3"
    "github.com/drep-project/drep-chain/transaction/types"
    "math/big"
)

type BlockHeader struct {
    ChainId              app.ChainIdType
    Version              int32
    PreviousHash         []byte
    GasLimit             *big.Int
    GasUsed              *big.Int
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
    TxList               []*types.Transaction
}

type Block struct {
    Header               *BlockHeader
    Data                 *BlockData
    MultiSig             *MultiSignature
}



type MultiSignature struct {
    Sig                  secp256k1.Signature
    Bitmap               []byte
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
package bean

import (
    "encoding/hex"
    "BlockChainTest/crypto"
)

func (tx *Transaction) GetId() (string, error) {
    b, err := Serialize(tx.Data)
    if err != nil {
        return "", err
    }
    id := hex.EncodeToString(crypto.Hash256(b))
    return id, nil
}

func (tx *Transaction) GetHash() ([]byte, error) {
    b, err := Serialize(tx)
    if err != nil {
        return nil, err
    }
    hash := crypto.Hash256(b)
    return hash, nil
}

func (block *Block) BlockID() (string, error) {
    b, err := Serialize(block.Header)
    if err != nil {
        return "", err
    }
    id := hex.EncodeToString(crypto.Hash256(b))
    return id, nil
}
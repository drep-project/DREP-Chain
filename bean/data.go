package bean

import (
    "encoding/hex"
    "BlockChainTest/hash"
    "github.com/golang/protobuf/proto"
)

func (tx *Transaction) GetId() (string, error) {
    b, err := proto.Marshal(tx.Data)
    if err != nil {
        return "", err
    }
    id := hex.EncodeToString(hash.Hash256(b))
    return id, nil
}

func (tx *Transaction) GetHash() ([]byte, error) {
    b, err := proto.Marshal(tx)
    if err != nil {
        return nil, err
    }
    h := hash.Hash256(b)
    return h, nil
}

func (block *Block) BlockID() (string, error) {
    b, err := proto.Marshal(block.Header)
    if err != nil {
        return "", err
    }
    id := hex.EncodeToString(hash.Hash256(b))
    return id, nil
}
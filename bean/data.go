package bean

import (
    "encoding/hex"
    "BlockChainTest/hash"
    "github.com/golang/protobuf/proto"
    "BlockChainTest/crypto"
)

func (tx *Transaction) TxId() (string, error) {
    b, err := proto.Marshal(tx.Data)
    if err != nil {
        return "", err
    }
    id := hex.EncodeToString(hash.Hash256(b))
    return id, nil
}

func (tx *Transaction) TxHash() ([]byte, error) {
    b, err := proto.Marshal(tx)
    if err != nil {
        return nil, err
    }
    h := hash.Hash256(b)
    return h, nil
}

func (tx *Transaction) TxSig() (*crypto.Signature, error) {
    b, err := proto.Marshal(tx.Data)
    if err != nil {
        return nil, err
    }
    return crypto.Sign(b)
}

func (block *Block) BlockID() (string, error) {
    b, err := proto.Marshal(block.Header)
    if err != nil {
        return "", err
    }
    id := hex.EncodeToString(hash.Hash256(b))
    return id, nil
}
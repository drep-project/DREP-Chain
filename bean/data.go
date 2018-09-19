package bean

import (
    "encoding/hex"
    "github.com/golang/protobuf/proto"
    "BlockChainTest/mycrypto"
)

const (
    AddressLen = 20
)

func (tx *Transaction) TxId() (string, error) {
    b, err := proto.Marshal(tx.Data)
    if err != nil {
        return "", err
    }
    id := hex.EncodeToString(mycrypto.Hash256(b))
    return id, nil
}

func (tx *Transaction) TxHash() ([]byte, error) {
    b, err := proto.Marshal(tx)
    if err != nil {
        return nil, err
    }
    h := mycrypto.Hash256(b)
    return h, nil
}

func (tx *Transaction) TxSig(prvKey *mycrypto.PrivateKey) (*mycrypto.Signature, error) {
    b, err := proto.Marshal(tx.Data)
    if err != nil {
        return nil, err
    }
    return mycrypto.Sign(prvKey, b)
}

func (tx *Transaction) Addr() Address {
    return Addr(tx.Data.PubKey)
}

func (block *Block) BlockID() (string, error) {
    b, err := proto.Marshal(block.Header)
    if err != nil {
        return "", err
    }
    id := hex.EncodeToString(mycrypto.Hash256(b))
    return id, nil
}

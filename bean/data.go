package bean

import (
    "encoding/hex"
    "github.com/golang/protobuf/proto"
    "BlockChainTest/mycrypto"
    "math/big"
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

func HeightToKey(height int64) string {
    return hex.EncodeToString(new(big.Int).SetInt64(height).Bytes())
}

func (block *Block) DBKey() string {
    return HeightToKey(block.Header.Height)
}

func (block *Block) DBMarshal() ([]byte, error) {
    _b, err := proto.Marshal(block)
    if err != nil {
        return nil, err
    }
    b := make([]byte, len(_b) + 1)
    b[0] = byte(MsgTypeBlock)
    copy(b[1:], _b)
    return b, nil
}
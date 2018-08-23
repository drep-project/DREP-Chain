package common

import (
    "bytes"
    "encoding/hex"
)

func (p *Point) Bytes() []byte {
    j := make([]byte, 2 * ByteLen)
    copy(j[ByteLen - len(p.X): ByteLen], p.X)
    copy(j[2 * ByteLen - len(p.Y):], p.Y)
    return j
}

func (p *Point) Equal(q *Point) bool {
    if !bytes.Equal(p.X, q.X) {
        return false
    }
    if !bytes.Equal(p.Y, q.Y) {
        return false
    }
    return true
}

func (pubKey *Point) Addr() string {
    j := pubKey.Bytes()
    h := Hash256(j)
    str := hex.EncodeToString(h[len(h) - AddressLen:])
    return str
}

func (tx *Transaction) TxID() (string, error) {
    b, err := Serialize(tx.Data)
    if err != nil {
        return "", err
    }
    id := hex.EncodeToString(Hash256(b))
    return id, nil
}

func (tx *Transaction) TxHash() ([]byte, error) {
    b, err := Serialize(tx)
    if err != nil {
        return nil, err
    }
    hash := Hash256(b)
    return hash, nil
}

func (block *Block) BlockID() (string, error) {
    b, err := Serialize(block.Header)
    if err != nil {
        return "", err
    }
    id := hex.EncodeToString(Hash256(b))
    return id, nil
}
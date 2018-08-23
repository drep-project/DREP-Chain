package common

import (
    "bytes"
    "encoding/hex"
    "errors"
    "math/big"
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

func (sig *Signature) Bytes() []byte {
    j := make([]byte, 2 * ByteLen)
    copy(j[ByteLen - len(sig.R): ByteLen], sig.R)
    copy(j[2 * ByteLen - len(sig.S):], sig.S)
    return j
}

func (tx *Transaction) TxConcat() ([]byte, error) {
    if tx.ToAddress == nil {
        return nil, errors.New("to address is nil")
    }
    len0 := len(tx.ToAddress)
    if tx.Amount == nil {
        return nil, errors.New("amount is nil")
    }
    len1 := len0 + len(tx.Amount)
    if tx.GasPrice == nil {
        return nil, errors.New("gas price is nil")
    }
    len2 := len1 + len(tx.GasPrice)
    if tx.GasLimit == nil {
        return nil, errors.New("gas limit is nil")
    }
    len3 := len2 + len(tx.GasLimit)
    bVersion := new(big.Int).SetInt64(int64(tx.Version)).Bytes()
    len4 := len3 + len(bVersion)
    bNonce := new(big.Int).SetInt64(tx.Nonce).Bytes()
    len5 := len4 + len(bNonce)
    bTimestamp := new(big.Int).SetInt64(tx.Timestamp).Bytes()
    len6 := len5 + len(bTimestamp)
    bPubKey := tx.PubKey.Bytes()
    len7 := len6 + len(bPubKey)
    concat := make([]byte, len7)
    copy(concat[:], tx.ToAddress)
    copy(concat[len0:], tx.Amount)
    copy(concat[len1:], tx.GasPrice)
    copy(concat[len2:], tx.GasLimit)
    copy(concat[len3:], bVersion)
    copy(concat[len4:], bNonce)
    copy(concat[len5:], bTimestamp)
    copy(concat[len6:], bPubKey)
    return concat, nil
}

func (tx *Transaction) TxID() (string, error) {
    concat, err := tx.TxConcat()
    if err != nil {
        return "", errors.New("concat transaction wrong")
    }
    id := hex.EncodeToString(Hash256(concat))
    return id, nil
}

func (tx *Transaction) TxHash() (*Transaction, error) {
    concat, err := tx.TxConcat()
    if err != nil {
        return nil, err
    }
    buf :=
}
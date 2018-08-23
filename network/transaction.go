package network

import (
    "bytes"
    "encoding/binary"
    "errors"
    "math/big"
)

var capacity = 1024 * 1024

func ToBytes(x interface{}) ([]byte, error) {
    buf := new(bytes.Buffer)
    err := binary.Write(buf, binary.BigEndian, x)
    if err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}

func CopyBytes(dst, src []byte, bgn, offset int) (int, error) {
    offset >>= 3
    if len(src) > offset {
        return bgn, errors.New("src too long")
    }
    if bgn + offset > cap(dst) {
        return bgn, errors.New("not enough room")
    }
    copy(dst[bgn + offset - len(src):], src)
    return bgn + offset, nil
}

func (tx *Transaction) GetTransactionMerge() ([]byte, error) {
    ret := make([]byte, capacity)
    n := 0

    // Version
    src, err := ToBytes(tx.Version)
    if err != nil {
        return nil, err
    }
    n, err = CopyBytes(ret, src, n, 32)
    if err != nil {
        return nil, err
    }

    // Nonce
    src, err = ToBytes(tx.Nonce)
    if err != nil {
        return nil, err
    }
    n, err = CopyBytes(ret, src, n, 64)
    if err != nil {
        return nil, err
    }

    // ToAddress
    n, err = CopyBytes(ret, tx.ToAddress, n, 160)
    if err != nil {
        return nil, err
    }

    // Amount
    n, err = CopyBytes(ret, tx.Amount, n, 128)
    if err != nil {
        return nil, err
    }

    // Gas Price
    n, err = CopyBytes(ret, tx.GasPrice, n, 128)
    if err != nil {
        return nil, err
    }

    // Gas Limit
    n, err = CopyBytes(ret, tx.GasLimit, n, 128)
    if err != nil {
        return nil, err
    }

    src, err = ToBytes(tx.Timestamp)
    if err != nil {
        return nil, err
    }
    n, err = CopyBytes(ret, src, n, 64)
    if err != nil {
        return nil, err
    }

    // PubKey
    n, err = CopyBytes(ret, tx.PubKey.X, n, 256)
    if err != nil {
        return nil, err
    }
    n, err = CopyBytes(ret, tx.PubKey.Y, n, 256)
    if err != nil {
        return nil, err
    }

    return ret[:n], nil
}

func (tx *Transaction) GetTransactionID() (string, error) {
    ret, err := tx.GetTransactionMerge()
    if err != nil {
        return "", err
    }
    id := new(big.Int).SetBytes(Hash256(ret)).String()
    return id, nil
}

func (tx *Transaction) GetTransactionHash() (string, error) {
    curve := InitCurve()
    prvKey, err := GenerateKey(curve)
    if err != nil {
        return "", err
    }
    ret, err := tx.GetTransactionMerge()
    if err != nil {
        return "", err
    }
    sig, err := Sign(curve, prvKey, ret)
    if err != nil {
        return "", err
    }
    tx.Sig = sig
    sigBytes := make([]byte, 512)
    copy(sigBytes, sig.R)
    _, err = CopyBytes(sigBytes, sig.S, 256, 256)
    if err != nil {
        return "", err
    }
    ret = append(ret, sigBytes...)
    id := new(big.Int).SetBytes(ret).String()
    return id, nil
}

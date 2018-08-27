package hash

import (
    "math/big"
    "crypto/sha256"
)

const (
    ByteLen = 32
)

func Hash256(text []byte) []byte {
    h := sha256.New()
    h.Write(text)
    ret := h.Sum(nil)
    hash := make([]byte, ByteLen)
    copy(hash[ByteLen - len(ret):], ret)
    return hash
}

func ConcatHash256(args ...[]byte) []byte {
    totalLen := 0
    for _, bytes := range args {
        totalLen += len(bytes)
    }
    concat := make([]byte, totalLen)
    i := 0
    for _, bytes := range args {
        copy(concat[i: ], bytes)
        i += len(bytes)
    }
    hash := Hash256(concat)
    return hash
}

func KDF(text []byte) []byte {
    pLen := len(text)
    k := pLen / ByteLen
    if pLen - k * ByteLen > 0 {
        k += 1
    }
    hash := make([]byte, k * ByteLen)
    p := new(big.Int).SetBytes(text)
    p.Lsh(p, 8)
    count := 0
    for count < k {
        c := new(big.Int).SetInt64(int64(count))
        b := new(big.Int).Add(p, c).Bytes()
        h := Hash256(b)
        copy(hash[count * ByteLen: (count + 1) * ByteLen], h)
        count += 1
    }
    return hash
}

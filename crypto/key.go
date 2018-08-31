package crypto

import (
    "bytes"
    "math/big"
    "encoding/hex"
    "BlockChainTest/hash"
)

const (
    ByteLen = 32
    AddressLen = 20
)

func (p *Point) Bytes() []byte {
    j := make([]byte, 2 * ByteLen)
    copy(j[ByteLen - len(p.X): ByteLen], p.Y)
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

func (p *Point) Int() (*big.Int, *big.Int) {
    return new(big.Int).SetBytes(p.X), new(big.Int).SetBytes(p.Y)
}

type Address string

func (addr Address) String() string {
    return string(addr)
}

func (pubKey *Point) Addr() Address {
    j := pubKey.Bytes()
    h := hash.Hash256(j)
    str := hex.EncodeToString(h[len(h) - AddressLen:])
    return Address(str)
}
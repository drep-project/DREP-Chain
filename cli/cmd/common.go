package cmd

import (
    "time"
    "math/rand"
    "encoding/hex"
)

const (
    AddressLength = 20
)

type CommonAddress [AddressLength]byte

func (addr CommonAddress) Bytes() []byte {
    return addr[:]
}

func (addr CommonAddress) Hex() string {
    return hex.EncodeToString(addr.Bytes())
}

func randomHex() string {
    seed := time.Now().Unix()
    source := rand.NewSource(seed)
    r := rand.New(source)
    p := make([]byte, AddressLength)
    r.Read(p)
    hexS := hex.EncodeToString(p)
    return hexS
}

package database

import (
    "testing"
    "math/big"
    "github.com/drep-project/drep-chain/crypto"
    "fmt"
)

var cache0 *Cache

func init() {
    cache0 = NewCache(db0)
}

func TestCacheBalance(t *testing.T) {
    addrBytes := []byte{0, 1, 2, 3, 4}
    addr := crypto.Bytes2Address(addrBytes)
    balance := new(big.Int).SetInt64(1000)
    cache0.putBalance(&addr, balance)
}

func TestGetCachedBalance(t *testing.T) {
    addrBytes := []byte{0, 1, 2, 3, 4}
    addr := crypto.Bytes2Address(addrBytes)
    fmt.Println("balance: ", cache0.getBalance(&addr))
}

func TestCacheNonce(t *testing.T) {
    addrBytes := []byte{0, 1, 2, 3, 4}
    addr := crypto.Bytes2Address(addrBytes)
    var nonce int64 = 1000
    cache0.putNonce(&addr, nonce)
}

func TestGetCachedNonce(t *testing.T) {
    addrBytes := []byte{0, 1, 2, 3, 4}
    addr := crypto.Bytes2Address(addrBytes)
    fmt.Println("balance: ", cache0.getNonce(&addr))
}

func TestCheckStateRoot(t *testing.T) {
    fmt.Println(cache0.GetRootValue())
}
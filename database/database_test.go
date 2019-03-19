package database

import (
    "testing"
    "fmt"
    "math/big"
    "github.com/drep-project/drep-chain/crypto"
)

var db *Database

func TestNewDatabase(t *testing.T) {
    var err error
    db, err = NewDatabase("newdb");
    if err != nil {
        fmt.Println(err)
    }
}

func TestPutBalance(t *testing.T) {
    addrBytes := []byte{0, 1, 2, 3, 4}
    addr := crypto.Bytes2Address(addrBytes)
    balance := new(big.Int).SetInt64(1000)
    fmt.Println("err: ", db.putBalance(&addr, balance))
}

func TestPutNonce(t *testing.T) {
    addrBytes := []byte{0, 1, 2, 3, 4}
    addr := crypto.Bytes2Address(addrBytes)
    var nonce int64 = 1000
    fmt.Println("err: ", db.putNonce(&addr, nonce))
}

func TestGetBalance(t *testing.T) {
    addrBytes := []byte{0, 1, 2, 3, 4}
    addr := crypto.Bytes2Address(addrBytes)
    fmt.Println("balance: ", db.getBalance(&addr))
}

func TestGetNonce(t *testing.T) {
    addrBytes := []byte{0, 1, 2, 3, 4}
    addr := crypto.Bytes2Address(addrBytes)
    fmt.Println("balance: ", db.getNonce(&addr))
}

func TestPutBlock(t *testing.T) {

}

func TestGetBlock(t *testing.T) {

}

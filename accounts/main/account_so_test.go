package main

import (
    "testing"
    "fmt"
)

func TestNewMain(t *testing.T) {
    prvKey, pubKey, chainCode, address := NewMainAccountKey()
    fmt.Println("private key: ", prvKey)
    fmt.Println("public key:  ", pubKey)
    fmt.Println("chain code:  ", chainCode)
    fmt.Println("address:     ", address)
    fmt.Println()

    msg := "34324343424ae22bf323"
    signature := Sign(prvKey, pubKey, msg)
    fmt.Println("signature:   ", signature)
    fmt.Println()

    ok := Verify(pubKey, msg, signature)
    fmt.Println("ok: ", ok)
    fmt.Println()
}

func TestNewSub(t *testing.T) {
    chainID := "3456"
    parentPrvKey := "c97b7e367d4474a89d5491054e7376c5f6b62cd8e3a73e3659c51c9ff226f5be"
    parentChainCode := "67d38def01931136f34f1cbf16b1d48dc83ea1077ac495f0db11e7ceb60a2fca"
    prvKey, pubKey, address := NewSubAccountKey(chainID, parentPrvKey, parentChainCode)
    fmt.Println("private key: ", prvKey)
    fmt.Println("public key:  ", pubKey)
    fmt.Println("address:     ", address)
    fmt.Println()

    msg := "34324343424ae22bf323"
    signature := Sign(prvKey, pubKey, msg)
    fmt.Println("signature:   ", signature)
    fmt.Println()

    ok := Verify(pubKey, msg, signature)
    fmt.Println("ok: ", ok)
    fmt.Println()
}
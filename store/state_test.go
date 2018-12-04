package store

import (
    "testing"
    "BlockChainTest/mycrypto"
    "fmt"
    "encoding/hex"
    "BlockChainTest/accounts"
)

func TestCalc(t *testing.T) {
    curve := mycrypto.GetCurve()
    k0 := []byte{0x22, 0x11}
    k1 := []byte{0x14, 0x44}
    k2 := []byte{0x11, 0x55}
    k3 := []byte{0x31, 0x63}
    pub0 := curve.ScalarBaseMultiply(k0)
    pub1 := curve.ScalarBaseMultiply(k1)
    pub2 := curve.ScalarBaseMultiply(k2)
    pub3 := curve.ScalarBaseMultiply(k3)
    fmt.Println("k0: ", hex.EncodeToString(k0))
    fmt.Println("k1: ", hex.EncodeToString(k1))
    fmt.Println("k2: ", hex.EncodeToString(k2))
    fmt.Println("k3: ", hex.EncodeToString(k3))
    fmt.Println("pub0: ", accounts.PubKey2Address(pub0).Hex())
    fmt.Println("pub1: ", accounts.PubKey2Address(pub1).Hex())
    fmt.Println("pub2: ", accounts.PubKey2Address(pub2).Hex())
    fmt.Println("pub3: ", accounts.PubKey2Address(pub3).Hex())
}

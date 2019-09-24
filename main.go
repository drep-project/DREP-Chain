package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
)

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
}

type AAA string

func main() {
	fmt.Println(hex.DecodeString("100000003000000031000000"))
}

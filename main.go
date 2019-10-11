package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"fmt"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/meling/urs"
	"math/rand"
	"time"
)

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
}

type AAA string

func main() {
	eatList := []string{"大娘水饺", "沙县小吃", "沙县隔壁", "河南面馆"}
	index := rand.Intn(int(time.Now().Unix()%int64(len(eatList))) + 1)
	fmt.Println(eatList[index])

	pri1, _ := urs.GenerateKey(secp256k1.S256(), crand.Reader)
	pri2, _ := urs.GenerateKey(secp256k1.S256(), crand.Reader)
	pr := urs.NewPublicKeyRing(2)
	pr.Add(pri1.PublicKey)
	pr.Add(pri2.PublicKey)

	xxxx, _ := urs.Sign(crand.Reader, pri1, pr, []byte{3, 4, 5})
	fmt.Print(xxxx)
	fmt.Println(urs.Verify(pr, []byte{3, 4, 5}, xxxx))

	xxxx2, _ := urs.Sign(crand.Reader, pri2, pr, []byte{3, 4, 5})
	fmt.Println(xxxx2)
	fmt.Print(urs.Verify(pr, []byte{3, 4, 5}, xxxx2))
}
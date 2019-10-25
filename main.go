package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/types"
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
	_,pk := secp256k1.PrivKeyFromScalar(common.MustDecode("0x29936977e507d150e2bb699042e605cdcebb59b83db372f8895b814904be8179"))
	dddd, _ := json.Marshal(pk)
	fmt.Println(string(dddd))
	fmt.Println(common.Encode(pk.Serialize()))
//	curl -H "Content-Type: application/json" -X post --data '{"jsonrpc":"2.0","method":"account_candidateCredit","params":["0x300fc5a14e578be28c64627c0e7e321771c58cd4","0x3641100","0x110","0x30000","{\"Pubkey\":\"0x03ad000bc9a4a29c11227d6b5ee05076b594c1c22567cdd85fbb8222e6a715924e\",\"Node\":\"enode://da388eb91ff35bc840b30a0adc771507cb5e933dff0818cf687644341e477e7e@192.168.147.134:55555\"}"],"id":1}' http://127.0.0.1:15645

str := "{\"Pubkey\":\"0x03ad000bc9a4a29c11227d6b5ee05076b594c1c22567cdd85fbb8222e6a715924e\",\"Node\":\"enode://da388eb91ff35bc840b30a0adc771507cb5e933dff0818cf687644341e477e7e@192.168.147.134:55555\"}"
	ddd := &types.CandidateData{}
	ec := 	ddd.Unmarshal([]byte(str))
	fmt.Println(ec)
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

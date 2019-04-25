package service

import (
	"crypto/rand"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"math/big"
	"testing"
)

func BenchmarkSigTransactionsSecp256k1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		privateKey, _ := crypto.GenerateKey(rand.Reader)
		pubkey := privateKey.PubKey()
		addr := crypto.PubKey2Address(pubkey)

		tx := chainTypes.NewTransaction(addr, crypto.CommonAddress{}, new(big.Int).SetInt64(100), uint64(i))
		secp256k1.SignCompact(privateKey, tx.TxHash().Bytes(), true)
	}
}
func BenchmarkVerifyTransactionsSecp256k1(b *testing.B) {
	cs := ChainService{}

	for i := 0; i < b.N; i++ {
		privateKey, _ := crypto.GenerateKey(rand.Reader)
		pubkey := privateKey.PubKey()
		addr := crypto.PubKey2Address(pubkey)

		tx := chainTypes.NewTransaction(addr, crypto.CommonAddress{}, new(big.Int).SetInt64(100), uint64(i))
		var err error
		tx.Sig, err = secp256k1.SignCompact(privateKey, tx.TxHash().Bytes(), true)
		bRet ,err := cs.verify(tx)
		if err != nil {
			b.Fatal(err)
		}
		if bRet == false{
			b.Fatal("verify error")
		}
	}
}


//func BenchmarkVerifyTransactionEcc(b *testing.B) {
//	//privateKey, _ := crypto.GenerateKey(rand.Reader)
//	//pubkey := privateKey.PubKey()
//	//addr := crypto.PubKey2Address(pubkey)
//	cs := ChainService{}
//
//	for i := 0; uint64(i) < 3000; i++ {
//		privateKey, _ := crypto.GenerateKey(rand.Reader)
//		pubkey := privateKey.PubKey()
//		addr := crypto.PubKey2Address(pubkey)
//
//		tx := chainTypes.NewTransaction(addr, crypto.CommonAddress{}, new(big.Int).SetInt64(100), uint64(i))
//		var err error
//		tx.Sig, err = secp256k1.SignCompact(privateKey, tx.TxHash().Bytes(), true)
//
//		bRet ,err := cs.verify(tx)
//		if err != nil {
//			b.Fatal(err)
//		}
//		if bRet == false{
//			b.Fatal("verify error")
//		}
//	}
//}

package txpool

import (
	"crypto/rand"

	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/types"

	"fmt"
	"math/big"
	"testing"
)

var j *txJournal

func rotateTx(t *testing.T, txs map[crypto.CommonAddress][]*types.Transaction) {
	err := j.rotate(txs)
	if err != nil {
		t.Fatal(err)
	}
}

func loadTx(t *testing.T, maxNonce uint64) {
	j = newTxJournal("./txpool/txs")
	err := j.load(func(txs []types.Transaction) []error {
		//for _, tx := range txs {
		//	fmt.Println(tx.Nonce(), tx.Amount(), tx.Type())
		//}
		if txs[len(txs)-1].Nonce() != maxNonce {
			return []error{fmt.Errorf("maybe nonce lost,%d != %d", txs[len(txs)-1].Nonce(), maxNonce)}
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

var generateMaxNonce uint64 = 10

func generateTxs() []*types.Transaction {
	privKey, _ := crypto.GenerateKey(rand.Reader)
	addr := crypto.PubKey2Address(privKey.PubKey())

	txs := make([]*types.Transaction, 0)

	for i := 0; i <= int(generateMaxNonce); i++ {
		amount := new(big.Int).SetUint64(100000000)
		tx := types.NewTransaction(addr, amount.Mul(amount, new(big.Int).SetUint64(uint64(i+1))), new(big.Int).SetUint64(100000000), new(big.Int).SetUint64(100000000), uint64(i))
		//fmt.Println(tx.Nonce(),tx.Amount(),tx.Type(),tx.GasPrice())
		sig, _ := secp256k1.SignCompact(privKey, tx.TxHash().Bytes(), true)
		tx.Sig = sig
		txs = append(txs, tx)
	}

	return txs
}

func TestLoadAndRotateNull(t *testing.T) {
	loadTx(t, 0)
	rotateTx(t, nil)
}

func TestLoadAndRotate(t *testing.T) {
	loadTx(t, 0)

	all := make(map[crypto.CommonAddress][]*types.Transaction)
	for j := 0; j < 1; j++ {
		txs := generateTxs()
		privateKey, _ := crypto.GenerateKey(rand.Reader)
		pubkey := privateKey.PubKey()
		addr := crypto.PubKey2Address(pubkey)
		all[addr] = txs
	}

	rotateTx(t, all)
}

var insertTxNum uint64 = 10

func insertTx(t *testing.T) {
	privKey, _ := crypto.GenerateKey(rand.Reader)
	addr := crypto.PubKey2Address(privKey.PubKey())

	for i := generateMaxNonce; i <= generateMaxNonce+insertTxNum; i++ {
		tx := types.NewTransaction(addr, new(big.Int).SetUint64(100000000), new(big.Int).SetUint64(100000000), new(big.Int).SetUint64(100000000), uint64(i))
		sig, _ := secp256k1.SignCompact(privKey, tx.TxHash().Bytes(), true)
		tx.Sig = sig
		err := j.insert(tx)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestInsert(t *testing.T) {
	loadTx(t, generateMaxNonce)

	rotateTx(t, nil)
	insertTx(t)

	loadTx(t, generateMaxNonce+insertTxNum)
}

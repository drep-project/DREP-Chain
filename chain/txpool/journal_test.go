package txpool

import (
	"crypto/rand"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"

	"fmt"
	"math/big"
	"testing"
)

var j *txJournal

func rotateTx(t *testing.T, txs map[crypto.CommonAddress][]*chainTypes.Transaction) {
	err := j.rotate(txs)
	if err != nil {
		t.Fatal(err)
	}
}

func loadTx(t *testing.T, maxNonce uint64) {
	j = newTxJournal("./txpool/txs")
	err := j.load(func(txs []chainTypes.Transaction) []error {
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

func generateTxs() []*chainTypes.Transaction {
	b := common.Bytes("0x0373654ccdb250f2cfcfe64c783a44b9ea85bc47f2f00c480d05082428d277d6d0")
	b.UnmarshalText(b)
	pubkey, _ := secp256k1.ParsePubKey(b)
	addr := crypto.PubKey2Address(pubkey)

	txs := make([]*chainTypes.Transaction, 0)

	for i := 0; i <= int(generateMaxNonce); i++ {
		amount := new(big.Int).SetUint64(100000000)
		tx := chainTypes.NewTransaction(addr, amount.Mul(amount, new(big.Int).SetUint64(uint64(i+1))), new(big.Int).SetUint64(100000000), new(big.Int).SetUint64(100000000), uint64(i))
		//fmt.Println(tx.Nonce(),tx.Amount(),tx.Type(),tx.GasPrice())
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

	all := make(map[crypto.CommonAddress][]*chainTypes.Transaction)
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
	b := common.Bytes("0x0373654ccdb250f2cfcfe64c783a44b9ea85bc47f2f00c480d05082428d277d6d0")
	b.UnmarshalText(b)
	pubkey, _ := secp256k1.ParsePubKey(b)
	addr := crypto.PubKey2Address(pubkey)

	for i := generateMaxNonce; i <= generateMaxNonce+insertTxNum; i++ {
		tx := chainTypes.NewTransaction(addr, new(big.Int).SetUint64(100000000), new(big.Int).SetUint64(100000000), new(big.Int).SetUint64(100000000), uint64(i))
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

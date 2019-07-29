package txpool

import (
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/types"

	rand2 "math/rand"

	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var txNum1 uint64 = 10000
var txNum2 int = 1000
var txPool *TransactionPool
var feed event.Feed

func TestNewTransactions(t *testing.T) {
	path := filepath.Join(os.TempDir(), fmt.Sprintf("./txpool/%d/data", rand2.Int63n(10000000)))
	diskDb, err := database.NewDatabase(path)
	if err != nil {
		t.Error("db init err")
	}

	path = filepath.Join(os.TempDir(), fmt.Sprintf("./jounal/txs"))
	txPool = NewTransactionPool(diskDb, path)
	if txPool == nil {
		t.Error("init database service err")
	}

	txPool.Start(&feed)
}

func addTx(t *testing.T, num uint64) error {
	privKey, _ := crypto.GenerateKey(rand.Reader)

	addr := crypto.PubKey2Address(privKey.PubKey())
	fmt.Println(string(addr.Hex()))
	txPool.database.BeginTransaction()

	var amount uint64 = 0xefffffffffffffff
	txPool.database.PutBalance(&addr, new(big.Int).SetUint64(amount))
	txPool.database.Commit(false)

	nonce := txPool.database.GetNonce(&addr)
	for i := 0; uint64(i) < num; i++ {
		tx := types.NewTransaction(addr, new(big.Int).SetInt64(100), new(big.Int).SetInt64(100), new(big.Int).SetInt64(100), nonce+uint64(i))

		sig, err := secp256k1.SignCompact(privKey, tx.TxHash().Bytes(), true)
		if err != nil {
			return err
		}

		tx.Sig = sig
		err = txPool.AddTransaction(tx, true)
		if err != nil {
			return err
		}
		if i%10000 == 0 {
			time.Sleep(time.Second * 1)
			fmt.Println("nonce id:", i)
		}
	}

	return nil
}

func TestAddTX(t *testing.T) {
	TestNewTransactions(t)
	err := addTx(t, txNum1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddIntevalTX(t *testing.T) {
	privKey, _ := crypto.GenerateKey(rand.Reader)
	addr := crypto.PubKey2Address(privKey.PubKey())
	for i := 0; i < txNum2; i++ {
		if i != 0 && i%100 == 0 {
			continue
		}

		tx := types.NewTransaction(addr, new(big.Int).SetUint64(100000000), new(big.Int).SetUint64(100000000), new(big.Int).SetUint64(100000000), uint64(i))
		txPool.AddTransaction(tx, true)
	}
}

//
//func TestSyncTx(t *testing.T) {
//	feed.Send(struct{}{})
//}

//池子里面的都是未处理的交易
func TestGetPendingTxs(t *testing.T) {
	TestNewTransactions(t)
	ch := make(chan uint64)
	go func() {
		err := addTx(t, txNum1)
		if err != nil {
			return
		}

		time.Sleep(time.Second * 1)
		ch <- txNum1
	}()

	func() {
		var nonce uint64
		for {
			select {
			case num := <-ch:
				if nonce != num {
					t.Fatalf("recv nonce:%d sendTxNum:%d", nonce, num)
				}
				return
			default:
				gasLimit := new(big.Int).SetInt64(10000000)
				mapTxs := txPool.GetPending(gasLimit)
				fmt.Println("pending tx len:", len(mapTxs))
				if len(mapTxs) == 0 {
					time.Sleep(time.Second * 1)
					continue
				}

				var addrs []*crypto.CommonAddress
				mapAddr := make(map[crypto.CommonAddress]struct{})
				for _, tx := range mapTxs {
					from := &crypto.CommonAddress{}
					if _, ok := mapAddr[*from]; !ok {
						mapAddr[*from] = struct{}{}
						addrs = append(addrs, from)
					}
					nonce = tx.Nonce()
				}
				fmt.Println("recv nonce:", nonce)
				txPool.database.BeginTransaction()
				txPool.database.Commit(false)

				feed.Send(addrs)
				time.Sleep(time.Second * 1)
			}
		}
	}()
}

//测试queue里面的tx被删除
func TestReplace(t *testing.T) {
	TestNewTransactions(t)

	privKey, _ := crypto.GenerateKey(rand.Reader)
	addr := crypto.PubKey2Address(privKey.PubKey())
	txPool.database.BeginTransaction()

	var amount uint64 = 0xefffffffffffffff
	txPool.database.PutBalance(&addr, new(big.Int).SetUint64(amount))
	txPool.database.Commit(false)

	nonce := txPool.database.GetNonce(&addr)
	for i := 0; uint64(i) < maxTxsOfPending; i++ {
		tx := types.NewTransaction(addr, new(big.Int).SetInt64(100), new(big.Int).SetInt64(int64(100+i)), new(big.Int).SetInt64(100), nonce+uint64(i))
		sig, err := secp256k1.SignCompact(privKey, tx.TxHash().Bytes(), true)
		tx.Sig = sig
		err = txPool.AddTransaction(tx, true)
		if err != nil {
			t.Fatal(err)
		}
	}

	nonce += maxTxsOfPending
	//20个到queue
	for i := 0; uint64(i) < maxTxsOfQueue; i++ {
		tx := types.NewTransaction(addr, new(big.Int).SetInt64(100), new(big.Int).SetInt64(int64(100+i+maxTxsOfPending)), new(big.Int).SetInt64(100), nonce+uint64(i))
		sig, err := secp256k1.SignCompact(privKey, tx.TxHash().Bytes(), true)
		tx.Sig = sig
		err = txPool.AddTransaction(tx, true)
		if err != nil {
			t.Fatal(err)
		}
	}

	nonce1 := nonce - 1
	//替换发生在pending
	for i := 0; uint64(i) < 1; i++ {
		tx := types.NewTransaction(addr, new(big.Int).SetInt64(100), new(big.Int).SetInt64(int64(100*4)), new(big.Int).SetInt64(100), nonce1+uint64(i))
		sig, err := secp256k1.SignCompact(privKey, tx.TxHash().Bytes(), true)
		tx.Sig = sig
		err = txPool.AddTransaction(tx, true)
		if err != nil {
			t.Fatal(err)
		}
	}

	nonce1 = nonce + 1
	//替换发生在queue
	for i := 0; uint64(i) < 1; i++ {
		tx := types.NewTransaction(addr, new(big.Int).SetInt64(100), new(big.Int).SetInt64(int64(100*4)), new(big.Int).SetInt64(100), nonce1+uint64(i))
		sig, err := secp256k1.SignCompact(privKey, tx.TxHash().Bytes(), true)
		tx.Sig = sig
		err = txPool.AddTransaction(tx, true)
		if err != nil {
			t.Fatal(err)
		}
	}
}

//测试pending里面tx被删除；同时删除导致nonce不连续，导致删除了多个tx
func TestDelTx(t *testing.T) {
	TestNewTransactions(t)

	privKey, _ := crypto.GenerateKey(rand.Reader)
	addr := crypto.PubKey2Address(privKey.PubKey())
	txPool.database.BeginTransaction()

	var amount uint64 = 0xefffffffffffffff
	txPool.database.PutBalance(&addr, new(big.Int).SetUint64(amount))
	txPool.database.Commit(false)

	nonce := txPool.getTransactionCount(&addr)
	for i := 0; uint64(i) < maxTxsOfQueue+maxTxsOfPending; i++ {
		tx := types.NewTransaction(addr, new(big.Int).SetInt64(100), new(big.Int).SetInt64(int64(100+i)), new(big.Int).SetInt64(100), nonce+uint64(i))
		sig, err := secp256k1.SignCompact(privKey, tx.TxHash().Bytes(), true)
		tx.Sig = sig
		err = txPool.AddTransaction(tx, false)
		if err != nil {
			t.Fatal(err)
		}
	}

	nonce += maxTxsOfQueue + maxTxsOfPending
	//删除发生在pending
	for i := 0; uint64(i) < 20; i++ {
		tx := types.NewTransaction(addr, new(big.Int).SetInt64(100), new(big.Int).SetInt64(int64(100*5)), new(big.Int).SetInt64(100), nonce+uint64(i))
		sig, err := secp256k1.SignCompact(privKey, tx.TxHash().Bytes(), true)
		tx.Sig = sig
		err = txPool.AddTransaction(tx, false)
		if err != nil {
			t.Fatal(err)
		}
	}
}

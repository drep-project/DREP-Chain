package database

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/drep-project/binary"
	"github.com/drep-project/DREP-Chain/crypto"
	chainTypes "github.com/drep-project/DREP-Chain/types"
)

func TestNewDatabase(t *testing.T) {

	os.RemoveAll("./test")

	//db, err := NewDatabase("./test")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//db.Close()

	_, err := NewDatabase("./test")
	if err != nil {
		t.Fatal(err)
	}

	os.RemoveAll("./test")
}

func TestAddLog(t *testing.T) {
	//defer os.RemoveAll("./test")
	//db, err := NewDatabase("./test")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//for i := 0; i < 100; i++ {
	//	pri, _ := crypto.GenerateKey(rand.Reader)
	//	addr := crypto.PubkeyToAddress(pri.PubKey())
	//
	//	log := chainTypes.Log{
	//		Address: addr,
	//		Height:  uint64(i),
	//		TxHash:  crypto.BytesToHash([]byte(strconv.Itoa(i))),
	//	}
	//
	//	err = db.AddLog(&log)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//}
	//
	//for i := 0; i < 100; i++ {
	//	log := db.GetLogs(crypto.BytesToHash([]byte(strconv.Itoa(i))))
	//	if log[0].Height != uint64(i) {
	//		t.Fatal("write log not equal read log")
	//	}
	//}
}

func TestBlockNode(t *testing.T) {
	defer os.RemoveAll("./test")
	db, err := NewDatabase("./test")
	if err != nil {
		t.Fatal(err)
	}

	hash := crypto.BytesToHash([]byte("hash"))
	//pri, _ := crypto.GenerateKey(rand.Reader)

	bn := chainTypes.BlockNode{
		Parent:    nil,
		Hash:      &hash,
		StateRoot: []byte{},
		TimeStamp: uint64(time.Now().Unix()),
		Height:    0,
		Status:    chainTypes.StatusInvalidAncestor,
		//LeaderPubKey: crypto.PubkeyToAddress(pri.PubKey()),
	}

	err = db.PutBlockNode(&bn)
	if err != nil {
		t.Fatal(err)
	}

	header, state, err := db.GetBlockNode(bn.Hash, bn.Height)
	if err != nil {
		t.Fatal(err)
	}

	if header.Height != bn.Height {
		t.Fatal("block node err")
	}

	if state.KnownValid() {
		t.Fatal("block state err")
	}
}

func TestNewTransaction(t *testing.T) {
	defer os.RemoveAll("./test")

	db, err := NewDatabase("./test")
	if err != nil {
		t.Fatal(err)
	}

	//数据提交
	db1 := db.BeginTransaction(true)
	root := db1.GetStateRoot()
	fmt.Println("01", db.GetStateRoot())

	pri, _ := crypto.GenerateKey(rand.Reader)
	addr := crypto.PubkeyToAddress(pri.PubKey())
	balance := new(big.Int).SetInt64(10000)

	db1.AddBalance(&addr, balance)
	db1.Commit()

	root1 := db1.GetStateRoot()
	if bytes.Equal(root, root1) {
		t.Fatal("root !=")
	}

	err = db.trieDb.Commit(crypto.Bytes2Hash(root1), false)
	if err != nil {
		t.Fatal(err)
	}

	balance = db1.GetBalance(&addr)

	fmt.Println(balance)

	db2 := db.BeginTransaction(true)
	root2 := db2.GetStateRoot()
	if !bytes.Equal(root1, root2) {
		t.Fatal("root !=")
	}

	fmt.Println("02", db2.GetStateRoot())
}

func TestDiscardCacheData(t *testing.T) {
	defer os.RemoveAll("./test")

	db, err := NewDatabase("./test")
	if err != nil {
		t.Fatal(err)
	}

	//数据不提交
	db2 := db.BeginTransaction(false)
	root := db2.GetStateRoot()

	pri, _ := crypto.GenerateKey(rand.Reader)
	addr := crypto.PubkeyToAddress(pri.PubKey())
	balance := new(big.Int).SetInt64(100001)
	db2.AddBalance(&addr, balance)

	root1 := db2.GetStateRoot()
	if !bytes.Equal(root, root1) {
		t.Fatal("root must equal")
	}

	db2.Commit()

	err = db.trieDb.Commit(crypto.Bytes2Hash(root1), false)
	if err != nil {
		t.Fatal(err)
	}

	db3 := db.BeginTransaction(false)
	root2 := db3.GetStateRoot()
	if !bytes.Equal(root, root2) {
		t.Fatal("root !=")
	}
}

func TestRecoverTrie(t *testing.T) {
	defer os.RemoveAll("./test")

	db, err := NewDatabase("./test")
	if err != nil {
		t.Fatal(err)
	}

	pri, _ := crypto.GenerateKey(rand.Reader)
	addr := crypto.PubkeyToAddress(pri.PubKey())
	balance := new(big.Int).SetInt64(100001)
	db2 := db.BeginTransaction(true)
	db2.AddBalance(&addr, balance)
	db2.Commit()

	root1 := db2.GetStateRoot()

	db2.AddBalance(&addr, balance)
	db2.Commit()

	root2 := db2.GetStateRoot()

	db.trieDb.Commit(crypto.Bytes2Hash(root2), false)

	b := db.RecoverTrie(root1)
	if !b {
		t.Fatal("recover trie err")
	}
}

func TestDatabase_UpdateCandidateAddr(t *testing.T) {
	defer os.RemoveAll("./test")

	db, err := NewDatabase("./test")
	if err != nil {
		t.Fatal(err)
	}

	b := db.RecoverTrie([]byte{})
	if b != true {
		t.Fatal("recover trie err")
	}

	db2 := db.BeginTransaction(true)
	address := make(map[crypto.CommonAddress]struct{})

	for i := 0; i < 10000; i++ {
		pri, _ := crypto.GenerateKey(rand.Reader)
		addr := crypto.PubkeyToAddress(pri.PubKey())
		address[addr] = struct{}{}

		err = db2.AddCandidateAddr(&addr)
		if err != nil {
			t.Fatal(err)
		}
		db2.Commit()
	}

	ret, err := db2.GetCandidateAddrs()
	if len(ret) != 10000 {
		t.Fatal("store err")
	}

	for addr, _ := range address {
		err = db2.DelCandidateAddr(&addr)
		if err != nil {
			t.Fatal(err)
		}
	}

	ret, err = db2.GetCandidateAddrs()
	if len(ret) > 0 {
		t.Fatal("store and del not match")
	}
}

func TestBinary(t *testing.T) {
	pri, _ := crypto.GenerateKey(rand.Reader)
	addr := crypto.PubkeyToAddress(pri.PubKey())

	m := make(map[crypto.CommonAddress]struct{})
	m[addr] = struct{}{}

	b, _ := binary.Marshal(&m)
	um := make(map[crypto.CommonAddress]struct{})
	err := binary.Unmarshal(b, &um)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := um[addr]; !ok {
		t.Fatal("unmarshal err")
	}
}

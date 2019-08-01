package database

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/drep-project/drep-chain/crypto"

	chainTypes "github.com/drep-project/drep-chain/types"
)

func TestNewDatabase(t *testing.T) {

	os.RemoveAll("./test")

	db, err := NewDatabase("./test")
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	_, err = NewDatabase("./test")
	if err != nil {
		t.Fatal(err)
	}

	os.RemoveAll("./test")
}

func TestAddLog(t *testing.T) {
	defer os.RemoveAll("./test")
	db, err := NewDatabase("./test")
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 100; i++ {
		pri, _ := crypto.GenerateKey(rand.Reader)
		addr := crypto.PubKey2Address(pri.PubKey())

		log := chainTypes.Log{
			Address: addr,
			Height:  uint64(i),
			TxHash:  crypto.BytesToHash([]byte(strconv.Itoa(i))),
		}

		err = db.AddLog(&log)
		if err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < 100; i++ {
		log := db.GetLogs(crypto.BytesToHash([]byte(strconv.Itoa(i))))
		if log[0].Height != uint64(i) {
			t.Fatal("write log not equal read log")
		}
	}
}

func TestBlockNode(t *testing.T) {
	defer os.RemoveAll("./test")
	db, err := NewDatabase("./test")
	if err != nil {
		t.Fatal(err)
	}

	hash := crypto.BytesToHash([]byte("hash"))
	pri, _ := crypto.GenerateKey(rand.Reader)

	bn := chainTypes.BlockNode{
		Parent:       nil,
		Hash:         &hash,
		StateRoot:    []byte{},
		TimeStamp:    uint64(time.Now().Unix()),
		Height:       0,
		Status:       chainTypes.StatusInvalidAncestor,
		LeaderPubKey: crypto.PubKey2Address(pri.PubKey()),
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
	addr := crypto.PubKey2Address(pri.PubKey())
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
	addr := crypto.PubKey2Address(pri.PubKey())
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

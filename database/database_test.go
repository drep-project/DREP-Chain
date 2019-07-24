package database

import (
	"crypto/rand"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"os"
	"strconv"
	"testing"
	"time"

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
			Height:  int64(i),
			TxHash:  crypto.BytesToHash([]byte(strconv.Itoa(i))),
		}

		err = db.AddLog(&log)
		if err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < 100; i++ {
		log := db.GetLogs(crypto.BytesToHash([]byte(strconv.Itoa(i))))
		if log[0].Height != int64(i) {
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
		Parent:    nil,
		Hash:      &hash,
		StateRoot: []byte{},
		TimeStamp: uint64(time.Now().Unix()),
		Height:    0,
		Status:    chainTypes.StatusInvalidAncestor,
		LeaderPubKey:secp256k1.PublicKey(pri.PublicKey),

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

func TestChainState(t*testing.T){
	defer os.RemoveAll("./test")
	db, err := NewDatabase("./test")
	if err != nil {
		t.Fatal(err)
	}

	bestState := chainTypes.BestState{
		Hash:crypto.Bytes2Hash([]byte("besthash")),
		PrevHash:crypto.Bytes2Hash([]byte("besthash")),
		Height:10,
	}

	db.PutChainState(&bestState)

	if err != nil {
		t.Fatal(err)
	}

	bs := db.GetChainState()
	if bestState.Height != bs.Height{
		t.Fatal("store state err")
	}
}

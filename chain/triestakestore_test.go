package chain

import (
	"crypto/rand"
	"fmt"
	"github.com/drep-project/drep-chain/crypto"
	"math/big"
	"os"
	"testing"
)

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

func TestVoteCredit(t *testing.T) {
	defer os.RemoveAll("./test")

	db, err := NewDatabase("./test")
	if err != nil {
		t.Fatal(err)
	}

	db1 := db.BeginTransaction(false)

	for i := 0; i < 100; i++ {
		pri, _ := crypto.GenerateKey(rand.Reader)
		addr := crypto.PubkeyToAddress(pri.PubKey())

		voteValue := new(big.Int).SetInt64(100000)
		db1.VoteCredit(&addr, &addr, voteValue)

		m := db1.GetVoteCredit(&addr)

		if v, ok := m[addr]; ok {
			if v.Cmp(voteValue) != 0 {
				t.Fatal("storage ang get not match")
			}
		} else {
			t.Fatal("storage value err")
		}
	}
}

func TestCancelVoteCredit(t *testing.T) {
	defer os.RemoveAll("./test")

	db, err := NewDatabase("./test")
	if err != nil {
		t.Fatal(err)
	}

	db1 := db.BeginTransaction(false)
	addrs := make([]crypto.CommonAddress, 0)

	for i := 0; i < 100; i++ {
		pri, _ := crypto.GenerateKey(rand.Reader)
		addr := crypto.PubkeyToAddress(pri.PubKey())

		voteValue := new(big.Int).SetInt64(100000)
		db1.VoteCredit(&addr, &addr, voteValue)
		addrs = append(addrs, addr)
	}

	for _, addr := range addrs {
		voteValue := new(big.Int).SetInt64(50000)
		err = db1.CancelVoteCredit(&addr, &addr, voteValue)
		if err != nil {
			t.Fatal("cancel vote ok")
		}

		m := db1.GetVoteCredit(&addr)
		if v, ok := m[addr]; ok {
			if v.Cmp(voteValue) != 0 {
				fmt.Println(v, voteValue)
				t.Fatal("storage ang get not match", v, voteValue)
			}
		} else {
			t.Fatal("storage value err")
		}
	}
}


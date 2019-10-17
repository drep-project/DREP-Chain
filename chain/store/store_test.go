package store

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/common/trie"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database/leveldb"
)

//添加 撤销
func TestGetVoteCredit(t *testing.T) {
	defer os.RemoveAll("./test")
	diskDB, _ := leveldb.New("./test", 16, 512, "")
	storeInterface, _ := TrieStoreFromStore(diskDB, trie.EmptyRoot[:])

	store := storeInterface.(*Store)

	b := store.RecoverTrie([]byte{})
	if b != true {
		t.Fatal("recover trie err")
	}

	pri, _ := crypto.GenerateKey(rand.Reader)
	backbone := crypto.PubkeyToAddress(pri.PubKey())

	total := new(big.Int)
	for i := 0; i < 10; i++ {
		pri, _ := crypto.GenerateKey(rand.Reader)
		addr := crypto.PubkeyToAddress(pri.PubKey())
		store.stake.VoteCredit(&addr, &backbone, new(big.Int).SetUint64(uint64(222+i)*drepUnit))
		total.Add(total, new(big.Int).SetUint64(uint64(222+i)*drepUnit))
	}

	store.stake.VoteCredit(&backbone, nil, new(big.Int).SetUint64(registerPledgeLimit*drepUnit))
	total.Add(total, new(big.Int).SetUint64(registerPledgeLimit*drepUnit))

	m, err := store.GetCandidateAddrs()
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := m[backbone]; !ok {
		t.Fatal("vote addr err")
	}

	if total.Cmp(store.GetVoteCreditCount(&backbone)) != 0 {
		t.Fatalf("vote coin err,%v,%v", total, store.GetVoteCreditCount(&backbone))
	}

	m1 := store.GetCreditDetails(&backbone)
	for addr, value := range m1 {
		fmt.Println(addr.String(), value)
	}
}

func TestPutBalance(t *testing.T) {
	defer os.RemoveAll("./test")
	diskDB, _ := leveldb.New("./test", 16, 512, "")
	storeInterface, _ := TrieStoreFromStore(diskDB, trie.EmptyRoot[:])

	store := storeInterface.(*Store)

	b := store.RecoverTrie([]byte{})
	if b != true {
		t.Fatal("recover trie err")
	}

	pri, _ := crypto.GenerateKey(rand.Reader)
	addr := crypto.PubkeyToAddress(pri.PubKey())
	err := store.PutBalance(&addr, 0, new(big.Int).SetUint64(10000))
	if err != nil {
		t.Fatal(err)
	}

	//todo + -
	store.stake.VoteCredit(&addr, nil, new(big.Int).SetUint64(222))
	store.stake.CancelVoteCredit(&addr, nil, new(big.Int).SetUint64(10), 0)

	err = store.AddBalance(&addr, ChangeCycle, new(big.Int).SetUint64(1111))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(store.GetBalance(&addr, 0))

	ret := new(big.Int).SetUint64(10000 + 10 + 1111).Cmp(store.GetBalance(&addr, ChangeCycle))
	if ret != 0 {
		t.Fatal("vote not merge to balance")
	}
}

func TestDatabase_UpdateCandidateAddr(t *testing.T) {
	defer os.RemoveAll("./test")

	diskDB, _ := leveldb.New("./test", 16, 512, "")
	storeInterface, err := TrieStoreFromStore(diskDB, trie.EmptyRoot[:])

	store := storeInterface.(*Store)

	b := store.RecoverTrie([]byte{})
	if b != true {
		t.Fatal("recover trie err")
	}

	address := make(map[crypto.CommonAddress]struct{})

	for i := 0; i < 1000; i++ {
		pri, _ := crypto.GenerateKey(rand.Reader)
		addr := crypto.PubkeyToAddress(pri.PubKey())
		address[addr] = struct{}{}

		err = store.stake.AddCandidateAddr(&addr)
		if err != nil {
			t.Fatal(err)
		}
	}

	ret, err := store.GetCandidateAddrs()
	if len(ret) != 10000 {
		t.Fatal("store err")
	}

	for addr, _ := range address {
		err = store.stake.DelCandidateAddr(&addr)
		if err != nil {
			t.Fatal(err)
		}
	}

	ret, err = store.GetCandidateAddrs()
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

	diskDB, _ := leveldb.New("./test", 16, 512, "")
	storeInterface, _ := TrieStoreFromStore(diskDB, trie.EmptyRoot[:])

	store := storeInterface.(*Store)

	b := store.RecoverTrie([]byte{})
	if b != true {
		t.Fatal("recover trie err")
	}

	for i := 0; i < 100; i++ {
		pri, _ := crypto.GenerateKey(rand.Reader)
		addr := crypto.PubkeyToAddress(pri.PubKey())

		voteValue := new(big.Int).SetInt64(100000)
		store.VoteCredit(&addr, &addr, voteValue)

		v := store.GetVoteCreditCount(&addr)
		if v.Cmp(voteValue) != 0 {
			t.Fatal("storage ang get not match")
		}
	}
}

func TestCancelVoteCredit(t *testing.T) {
	diskDB, _ := leveldb.New("./test", 16, 512, "")
	storeInterface, err := TrieStoreFromStore(diskDB, trie.EmptyRoot[:])

	store := storeInterface.(*Store)

	b := store.RecoverTrie([]byte{})
	if b != true {
		t.Fatal("recover trie err")
	}

	addrs := make([]crypto.CommonAddress, 0)

	for i := 0; i < 100; i++ {
		pri, _ := crypto.GenerateKey(rand.Reader)
		addr := crypto.PubkeyToAddress(pri.PubKey())

		voteValue := new(big.Int).SetInt64(100000)
		store.VoteCredit(&addr, &addr, voteValue)
		addrs = append(addrs, addr)
	}

	for _, addr := range addrs {
		voteValue := new(big.Int).SetInt64(50000)
		err = store.CancelVoteCredit(&addr, &addr, voteValue, 0)
		if err != nil {
			t.Fatal("cancel vote ok")
		}

		v := store.GetVoteCreditCount(&addr)

		if v.Cmp(voteValue) != 0 {
			fmt.Println(v, voteValue)
			t.Fatal("storage ang get not match", v, voteValue)
		}

	}
}

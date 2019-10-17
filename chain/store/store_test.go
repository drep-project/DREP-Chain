package store

import (
	"crypto/rand"
	"fmt"
	"github.com/drep-project/drep-chain/types"
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

	if total.Cmp(store.GetVoteCreditCount(&backbone)) != 0 {
		t.Fatalf("vote coin err,%v,%v", total, store.GetVoteCreditCount(&backbone))
	}

	m1 := store.GetCreditDetails(&backbone)
	for addr, value := range m1 {
		fmt.Println(addr.String(), value)
	}
}

func TestCandidateCredit(t *testing.T) {
	defer os.RemoveAll("./test")
	diskDB, _ := leveldb.New("./test", 16, 512, "")
	storeInterface, _ := TrieStoreFromStore(diskDB, trie.EmptyRoot[:])

	pri, _ := crypto.GenerateKey(rand.Reader)
	backbone := crypto.PubkeyToAddress(pri.PubKey())

	store := storeInterface.(*Store)
	pk, _ := crypto.GenerateKey(rand.Reader)
	pkText, _ := pk.PubKey().MarshalText()
	cd := &types.CandidateData{
		P2PPubkey: string(pkText),
		Addr:      "127.0.0.1:55555",
	}
	data, _ := cd.Marshal()
	store.stake.CandidateCredit(&backbone, new(big.Int).SetUint64(registerPledgeLimit*drepUnit), data)

	m, err := store.GetCandidateAddrs()
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := m[backbone]; !ok {
		t.Fatal("vote addr err")
	}

	m1 := store.GetCreditDetails(&backbone)
	for addr, value := range m1 {
		fmt.Println(addr.String(), value)
	}

	b, _ := store.GetCandidateData(&backbone)
	fmt.Println(string(b))
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
	backbone := crypto.PubkeyToAddress(pri.PubKey())

	pri, _ = crypto.GenerateKey(rand.Reader)
	addr := crypto.PubkeyToAddress(pri.PubKey())
	err := store.PutBalance(&addr, 0, new(big.Int).SetUint64(10000))
	if err != nil {
		t.Fatal(err)
	}

	//todo + -
	store.stake.VoteCredit(&addr, &backbone, new(big.Int).SetUint64(100))
	store.stake.CancelVoteCredit(&addr, &backbone, new(big.Int).SetUint64(10), 0)

	err = store.AddBalance(&addr, ChangeCycle, new(big.Int).SetUint64(20))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(store.GetBalance(&addr, 0))

	ret := new(big.Int).SetUint64(10000 + 10 + 20).Cmp(store.GetBalance(&addr, ChangeCycle))
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
	if len(ret) != 1000 {
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

	pri, _ := crypto.GenerateKey(rand.Reader)
	backbone := crypto.PubkeyToAddress(pri.PubKey())

	for i := 0; i < 100; i++ {
		pri, _ := crypto.GenerateKey(rand.Reader)
		addr := crypto.PubkeyToAddress(pri.PubKey())

		voteValue := new(big.Int).SetInt64(100000)
		store.VoteCredit(&addr, &backbone, voteValue)

		v := store.GetVoteCreditCount(&backbone)
		voteValue = new(big.Int).SetUint64(uint64(100000 * (i + 1)))
		if v.Cmp(voteValue) != 0 {
			t.Fatalf("storage ang get not match,%v != %v", v, voteValue)
		}
	}
}

func TestCancelVoteCredit(t *testing.T) {
	defer os.RemoveAll("./test")

	diskDB, _ := leveldb.New("./test", 16, 512, "")
	storeInterface, err := TrieStoreFromStore(diskDB, trie.EmptyRoot[:])

	store := storeInterface.(*Store)

	b := store.RecoverTrie([]byte{})
	if b != true {
		t.Fatal("recover trie err")
	}

	pri, _ := crypto.GenerateKey(rand.Reader)
	backbone := crypto.PubkeyToAddress(pri.PubKey())
	addrs := make([]crypto.CommonAddress, 0)

	for i := 0; i < 10; i++ {
		pri, _ := crypto.GenerateKey(rand.Reader)
		addr := crypto.PubkeyToAddress(pri.PubKey())

		voteValue := new(big.Int).SetInt64(100000)
		store.VoteCredit(&addr, &backbone, voteValue)
		addrs = append(addrs, addr)
	}

	for _, addr := range addrs {
		voteValue := new(big.Int).SetInt64(50000)
		err = store.CancelVoteCredit(&addr, &backbone, voteValue, 0)
		if err != nil {
			t.Fatal("cancel vote ok")
		}

	}

	v := store.GetVoteCreditCount(&backbone)

	if v.Cmp(new(big.Int).SetUint64(100000*10/2)) != 0 {
		fmt.Println(v)
		t.Fatal("storage ang get not match", v, 100000*10/2)
	}

	m := store.GetCreditDetails(&backbone)
	for k, v := range m {
		fmt.Println(k.String(), v)
	}

	for _, addr := range addrs {
		b := store.GetBalance(&addr, ChangeCycle)
		if b.Cmp(new(big.Int).SetInt64(50000)) != 0 {
			t.Fatalf("cancel vote err,%v", b)
		}
	}

	for _, addr := range addrs {
		b := store.GetBalance(&addr, 0)
		if b.Cmp(new(big.Int).SetInt64(0)) != 0 {
			t.Fatalf("cancel vote err,%v", b)
		}
	}
}

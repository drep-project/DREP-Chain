package database

import (
	"bytes"
	"fmt"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database/drepdb/memorydb"
	"github.com/drep-project/drep-chain/database/trie"
	"strconv"
	"testing"
)

func TestGetAndSet(t *testing.T) {
	diskDB := memorydb.New()
	trieDB := trie.NewDatabase(diskDB)
	tree, _ := trie.NewSecure(crypto.Hash{}, trieDB)
	cacheStore := NewTransactionStore(tree, diskDB)

	key := []byte("1011")
	value := []byte("value-test")

	err := cacheStore.Put(key, value)
	if err != nil {
		t.Fatal(err)
	}

	b, err := cacheStore.Get([]byte(key))
	if err != nil || bytes.Compare(b, value) != 0 {
		t.Fatal(err)
	}
}

func TestFlushAndGet(t *testing.T) {
	diskDB := memorydb.New()
	trieDB := trie.NewDatabase(diskDB)
	tree, _ := trie.NewSecure(crypto.Hash{}, trieDB)
	cacheStore := NewTransactionStore(tree, diskDB)

	for i := 0; i < 100; i++ {
		value := []byte(strconv.Itoa(i) + "-value-test")
		if i%10 == 0 {
			value = []byte{}
		}
		key := []byte(strconv.Itoa(i))

		err := cacheStore.Put(key, value)
		if err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < 100; i++ {
		key := []byte(strconv.Itoa(i))
		value, err := cacheStore.trie.TryGet(key)
		if err != nil {
			t.Fatal(err)
		}
		if string(value) != "" {
			fmt.Println(value)
		}
	}

	cacheStore.Flush(false)

	for i := 0; i < 100; i++ {
		key := []byte(strconv.Itoa(i))
		value := []byte(strconv.Itoa(i) + "-value-test")
		v, err := cacheStore.Get(key)
		if err != nil || bytes.Compare(value, v) != 0 {
			t.Fatal(err)
		}
	}

	for i := 0; i < 100; i++ {
		key := []byte(strconv.Itoa(i))
		value := []byte(strconv.Itoa(i) + "-value-test")
		v, err := cacheStore.trie.TryGet(key)
		if err != nil || bytes.Compare(value, v) != 0 {
			t.Fatal(err)
		}
	}
}

func TestDelete(t *testing.T) {
	diskDB := memorydb.New()
	trieDB := trie.NewDatabase(diskDB)
	tree, _ := trie.NewSecure(crypto.Hash{}, trieDB)
	cacheStore := NewTransactionStore(tree, diskDB)

	for i := 0; i < 100; i++ {
		key := []byte(strconv.Itoa(i))
		value := []byte(strconv.Itoa(i) + "-value-test")
		err := cacheStore.Put(key, value)
		if err != nil {
			t.Fatal(err)
		}
	}

	key := []byte(strconv.Itoa(1))
	cacheStore.Delete(key)
}

func TestCopyState(t *testing.T) {

	diskDB := memorydb.New()
	trieDB := trie.NewDatabase(diskDB)
	tree, _ := trie.NewSecure(crypto.Hash{}, trieDB)
	cacheStore := NewTransactionStore(tree, diskDB)

	for i := 0; i < 100; i++ {
		key := []byte(strconv.Itoa(i))
		value := []byte(strconv.Itoa(i) + "-value-test")
		err := cacheStore.Put(key, value)
		if err != nil {
			t.Fatal(err)
		}
	}

	m := cacheStore.CopyState()
	for i := 0; i < 100; i++ {
		key := []byte(strconv.Itoa(i))
		if _, ok := m.Load(key); !ok {
			t.Fatal("copy err")
		}
	}

}

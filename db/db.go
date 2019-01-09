package db

import (
	"github.com/syndtr/goleveldb/leveldb"
	"BlockChainTest/trie"
	"BlockChainTest/config"
	"fmt"
	"BlockChainTest/util/list"
	"bytes"
)


func (db *Database) GetTotalStateRoot() []byte {
	if db.runningChain != config.RootChain {
		return db.GetChainStateRoot(db.runningChain)
	}
	type trieObj struct {
		chainId config.ChainIdType
		tr *trie.StateTrie
	}
	sll := list.NewSortedLinkedList(func(a interface{}, b interface{}) int {
		ac := a.(*trieObj).chainId
		bc := b.(*trieObj).chainId
		return bytes.Compare(ac[:], bc[:])
	})
	for chainId, t := range db.tries {
		sll.Add(&trieObj{
			chainId: chainId,
			tr: t,
		})
	}
	ts := make([]*trie.StateTrie, sll.Size())
	for i, elem := range sll.ToArray() {
		ts[i] = elem.(*trieObj).tr
	}
	return trie.GetMerkleRoot(ts)
}

func (db *Database) GetChainStateRoot(chainId config.ChainIdType) []byte {
	if t, exists := db.tries[chainId]; exists {
		return t.Root.Value
	} else {
		return nil
	}
}
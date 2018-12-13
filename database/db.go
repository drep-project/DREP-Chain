package database

import (
	"github.com/syndtr/goleveldb/leveldb"
	"BlockChainTest/trie"
	"BlockChainTest/config"
	"fmt"
	"BlockChainTest/util/list"
	"bytes"
)

const (
	ins = iota
	mod
	del
)

type journalEntry struct {
	chainId config.ChainIdType
	onTrie  bool
	action  int
	key     []byte
	prev    []byte
}

type Transactional interface {
	Put(chainId config.ChainIdType, key []byte, value []byte)
	Get(key []byte) []byte
	Delete(chainId config.ChainIdType, key []byte)
	Commit()
	Discard()
	BeginTransaction() Transactional
	GetTotalStateRoot() []byte
	GetChainStateRoot(chainId config.ChainIdType) []byte
}

type Transaction struct {
	parent   Transactional
	finished bool
	journal  []*journalEntry
}

func (t *Transaction) Put(chainId config.ChainIdType, key []byte, value []byte) {
	if t.finished {
		return
	}
	prev := t.parent.Get(key)
	if prev == nil {
		t.journal = append(t.journal, &journalEntry{chainId: chainId, action: ins, key: key})
	} else {
		t.journal = append(t.journal, &journalEntry{chainId: chainId, action: mod, key: key, prev:prev})
	}
	t.parent.Put(chainId, key, value)
}

func (t *Transaction) Get(key []byte) []byte {
	if t.finished {
		return nil
	}
	return t.parent.Get(key)
}

func (t *Transaction) Delete(chainId config.ChainIdType, key []byte) {
	if t.finished {
		return
	}
	prev := t.parent.Get(key)
	if prev == nil {
		return
	}
	t.journal = append(t.journal, &journalEntry{chainId:chainId, action: del, key:key, prev:prev})
	//if t.database.tries[chainId] == nil {
	//	t.database.tries[chainId] = trie.NewStateTrie()
	//}
	//t.database.tries[chainId].Delete(key)
	t.parent.Delete(chainId, key)
}

func (t *Transaction) Commit() {
	if t.finished {
		return
	}
	t.finished = true
}

func (t *Transaction) Discard() {
	if t.finished {
		return
	}
	t.finished = true
	for i := len(t.journal); i >= 0; i-- {
		e := t.journal[i]
		switch e.action {
		case ins:
			t.parent.Delete(e.chainId, e.key)
		case mod:
			t.parent.Put(e.chainId, e.key, e.prev)
		case del:
			t.parent.Put(e.chainId, e.key, e.prev)
		}
	}
	//for _, j := range t.journal {
	//
	//	switch j.action {
	//	case del:
	//		chainId := j.chainId
	//		if t.database.tries[chainId] == nil {
	//			t.database.tries[chainId] = trie.NewStateTrie()
	//		}
	//		if value, err := t.snapshot.Get(j.key, nil); err == nil {
	//			t.database.tries[chainId].Insert(j.key, value)
	//		}
	//	case put:
	//		chainId := j.chainId
	//		if t.database.tries[chainId] == nil {
	//			t.database.tries[chainId] = trie.NewStateTrie()
	//		}
	//		if value, err := t.snapshot.Get(j.key, nil); err == nil {
	//			t.database.tries[chainId].Insert(j.key, value)
	//		} else if err == leveldb.ErrNotFound {
	//			t.database.tries[chainId].Delete(j.key)
	//		}
	//	}
	//}
}

func (t *Transaction) BeginTransaction() Transactional {
	return &Transaction{
		parent:t,
		finished:false,
		journal:make([]*journalEntry, 0),
	}
}

func (t *Transaction) GetTotalStateRoot() []byte {
	return t.parent.GetTotalStateRoot()
}

func (t *Transaction) GetChainStateRoot(chainId config.ChainIdType) []byte {
	return t.parent.GetChainStateRoot(chainId)
}

type Database struct {
	db           *leveldb.DB
	runningChain config.ChainIdType
	rootChain    config.ChainIdType
	tries        map[config.ChainIdType] *trie.StateTrie
}

func NewDatabase(cfg *config.NodeConfig) *Database {
	ldb, err := leveldb.OpenFile(cfg.DbPath, nil)
	if err != nil {
		return nil
	}
	return &Database{
		db:ldb,
		runningChain: cfg.ChainId,
		tries: make(map[config.ChainIdType] *trie.StateTrie),
	}
}

func (db *Database) Put(chainId config.ChainIdType, key []byte, value []byte) {
	if err := db.db.Put(key, value, nil); err == nil {
		t, exists := db.tries[chainId]
		if !exists {
			t = trie.NewStateTrie()
			db.tries[chainId] = t
		}
		t.Insert(key, value)
	} else {
		fmt.Println("error occurs", err)
	}
}


func (db *Database) Get(key []byte) []byte {
	if ret, err := db.db.Get(key, nil); err == nil {
		return ret
	} else {
		return nil
	}
}

func (db *Database) Delete(chainId config.ChainIdType, key []byte) {
	if err := db.db.Delete(key, nil); err == nil {
		t, exists := db.tries[chainId]
		if !exists {
			fmt.Println("What the fuck, the trie dose not exist.")
			t = trie.NewStateTrie()
			db.tries[chainId] = t
		}
		t.Delete(key)
	} else {
		fmt.Println("Error occurs.", err)
	}
}

func (db *Database) Commit() {
}

func (db *Database) Discard() {
}

func (db *Database) BeginTransaction() Transactional {
	return &Transaction{
		parent:   db,
		finished: false,
		journal:  make([]*journalEntry, 0),
	}
}

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

// TODO 下面的函数操作不修改trie

func (db *Database) PutNoTrie(key []byte, value []byte) {
	if err := db.db.Put(key, value, nil); err != nil {
		fmt.Println("error occurs", err)
	}
}

func (db *Database) DeleteNoTrie(key []byte) {
	if err := db.db.Delete(key, nil); err != nil {
		fmt.Println("Error occurs.", err)
	}
}
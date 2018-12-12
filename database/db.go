package database

import (
	"github.com/syndtr/goleveldb/leveldb"
	"BlockChainTest/trie"
	"BlockChainTest/util/list"
	"encoding/hex"
)

type Database struct {
	db *leveldb.DB
	runningChain int64
	tries map[int64] *trie.StateTrie
}

const (
	ins = iota
	mod
	del
)

type journalEntry struct {
	chainId int64
	onTrie bool
	action int
	key []byte
	prev []byte
}

type Transactional interface {
	Put(chainId int64, key []byte, value []byte)
	Get(key []byte) []byte
	Delete(chainId int64, key []byte)
	Commit()
	Discard()
	BeginTransaction() *Transactional
	insertTrie(chainId int64, key []byte, value []byte)
	deleteTrie(chainId int64, key []byte)
}

type Transaction struct {
	parent   Transactional
	finished bool
	journal  []*journalEntry
}

func (t *Transaction) Put(chainId int64, key []byte, value []byte) {
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
	t.parent.insertTrie(chainId, key, value)
	//if t..tries[chainId] == nil {
	//	t.database.tries[chainId] = trie.NewStateTrie()
	//}
	//t.database.tries[chainId].Insert(key, value)
}

func (t *Transaction) Get(key []byte) []byte {
	if t.finished {
		return nil
	}
	return t.parent.Get(key)
	//} else if err == leveldb.ErrNotFound{
	//	return nil
	//} else {
	//	return nil
	//}
}

func (t *Transaction) Delete(key []byte, chainId int64, onTrie bool) {
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
	t.parent.deleteTrie(chainId, key)
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
			t.parent.deleteTrie(e.chainId, e.key)
		case mod:
			t.parent.Put(e.chainId, e.key, e.prev)
			t.parent.insertTrie(e.chainId, e.key, e.prev)
		case del:
			t.parent.Put(e.chainId, e.key, e.prev)
			t.parent.insertTrie(e.chainId, e.key, e.prev)
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

func (t *Transaction) BeginTransaction() *Transactional {
	return t
}

func NewDatabase(config *config.NodeConfig) *Database {
	ldb, err := leveldb.OpenFile(config.DbPath, nil)
	if err != nil {
		return nil
	}
	return &Database{
		db:ldb,
		runningChain: config.ChainId,
		tries: make(map[int64] *trie.StateTrie),
	}
}

func (db *Database) BeginTransaction() *Transaction {
	if s, err := db.db.GetSnapshot(); err == nil {
		return &Transaction{
			database: db,
			snapshot: s,
			finished: false,
			journals: make([]*journal, 0),
			values:   make(map[string][]byte),
		}
	} else {
		return nil
	}
}

func (db *Database) GetStateRoot() []byte {
	if db.runningChain != config.RootChain {
		return db.tries[db.runningChain].Root.Value
	}
	type trieObj struct {
		chainId int64
		tr *trie.StateTrie
	}
	sll := list.NewSortedLinkedList(func(a interface{}, b interface{}) int {
		ac := a.(*trieObj).chainId
		bc := b.(*trieObj).chainId
		if ac > bc {
			return 1
		}//tps
		if ac == bc {
			return 0
		}
		return -1
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

func (db *Database) put(key []byte, value []byte, chainId int64) error {
	err := db.db.Put(key, value, nil)
	fmt.Println("db put, key: ", hex.EncodeToString(key))
	if err == nil {
		if db.tries[chainId] == nil {
			db.tries[chainId] = trie.NewStateTrie()
		}
		db.tries[chainId].Insert(key, value)
	}
	return err
}

func (db *Database) get(key []byte) []byte {
	if ret, err := db.db.Get(key, nil); err == nil {
		return ret
	} else {
		return nil
	}
}
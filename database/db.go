package database

import (
	"github.com/syndtr/goleveldb/leveldb"
	"BlockChainTest/trie"
	"BlockChainTest/config"
	"fmt"
	"BlockChainTest/util/list"
	"encoding/hex"
)

type Database struct {
	db *leveldb.DB
	runningChain int64
	tries map[int64] *trie.StateTrie
}

const (
	del = iota
	put
)

type journal struct {
	chainId int64
	onTrie bool
	action int
	key []byte
	value []byte
}

type Transaction struct {
	database *Database
	snapshot *leveldb.Snapshot
	finished bool
	journals []*journal
	values map[string][]byte
}

func (t *Transaction) Put(key []byte, value []byte, chainId int64, onTrie bool) {
	fmt.Println("transaction put, key: ", hex.EncodeToString(key))
	if t.finished {
		return
	}
	t.journals = append(t.journals, &journal{
		chainId: chainId,
		onTrie: onTrie,
		action: put,
		key: key,
		value: value,
	})
	t.values[string(key)] = value
	if !onTrie {
		return
	}
	if t.database.tries[chainId] == nil {
		t.database.tries[chainId] = trie.NewStateTrie()
	}
	t.database.tries[chainId].Insert(key, value)
	//fmt.Println()
	//fmt.Println("key: ", hex.EncodeToString(key))
	//fmt.Println("value: ", hex.EncodeToString(value))
	//fmt.Println("state root: ", hex.EncodeToString(GetStateRoot()))
	//fmt.Println()
}

func (t *Transaction) Get(key []byte) []byte {
	if t.finished {
		return nil
	}
	if value, exists := t.values[string(key)]; exists {
		return value
	} else if value, err := t.snapshot.Get(key, nil); err == nil {
		return value
	} else if err == leveldb.ErrNotFound{
		return nil
	} else {
		return nil
	}
}

func (t *Transaction) Delete(key []byte, chainId int64, onTrie bool) {
	if t.finished {
		return
	}
	t.journals = append(t.journals, &journal{
		chainId: chainId,
		onTrie: onTrie,
		action: del,
		key: key,
	})
	if !onTrie {
		return
	}
	if t.database.tries[chainId] == nil {
		t.database.tries[chainId] = trie.NewStateTrie()
	}
	t.database.tries[chainId].Delete(key)
	delete(t.values, string(key))
}

func (t *Transaction) Commit() {
	if t.finished {
		return
	}
	t.finished = true
	tran, err := t.database.db.OpenTransaction()
	if err != nil {
		fmt.Errorf("error occurs when opening transaction: %v\n", err)
		return
	}
	for _, j := range t.journals {
		switch j.action {
		case del:
			if err := tran.Delete(j.key, nil); err != nil {
				fmt.Errorf("error occurs when deleting: %v\n", err)
			}
		case put:
			if err := tran.Put(j.key, j.value, nil); err != nil {
				fmt.Errorf("error occurs when puting data: %v\n", err)
			}
		}
	}
	if err := tran.Commit(); err != nil {
		fmt.Errorf("error occurs when committing: %v\n", err)
	}
}

func (t *Transaction) Discard() {
	if t.finished {
		return
	}
	t.finished = true
	for _, j := range t.journals {
		if !j.onTrie {
			continue
		}
		switch j.action {
		case del:
			chainId := j.chainId
			if t.database.tries[chainId] == nil {
				t.database.tries[chainId] = trie.NewStateTrie()
			}
			if value, err := t.snapshot.Get(j.key, nil); err == nil {
				t.database.tries[chainId].Insert(j.key, value)
			}
		case put:
			chainId := j.chainId
			if t.database.tries[chainId] == nil {
				t.database.tries[chainId] = trie.NewStateTrie()
			}
			if value, err := t.snapshot.Get(j.key, nil); err == nil {
				t.database.tries[chainId].Insert(j.key, value)
			} else if err == leveldb.ErrNotFound {
				t.database.tries[chainId].Delete(j.key)
			}
		}
	}
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
		}
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
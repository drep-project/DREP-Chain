package database

import (
	"github.com/syndtr/goleveldb/leveldb"
	"BlockChainTest/trie"
	"fmt"
)

type Database struct {
	db *leveldb.DB
	trie    *trie.StateTrie
}

const (
	del = iota
	put
)

type journal struct {
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

func (t *Transaction) Put(key []byte, value []byte) {
	if t.finished {
		return
	}
	t.journals = append(t.journals, &journal{action:put, key:key, value:value})
	t.values[string(key)] = value
	t.database.trie.Insert(key, value)
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

func (t *Transaction) Delete(key []byte) {
	if t.finished {
		return
	}
	t.journals = append(t.journals, &journal{action:del, key:key})
	t.database.trie.Delete(key)
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
		switch j.action {
		case del:
			if value, err := t.snapshot.Get(j.key, nil); err == nil {
				t.database.trie.Insert(j.key, value)
			}
		case put:
			if value, err := t.snapshot.Get(j.key, nil); err == nil {
				t.database.trie.Insert(j.key, value)
			} else if err == leveldb.ErrNotFound {
				t.database.trie.Delete(j.key)
			}
		}
	}
}

const databaseName = "local_data"

func NewDatabase() *Database {
	ldb, err := leveldb.OpenFile(databaseName, nil)
	if err != nil {
		return nil
	}
	return &Database{db:ldb, trie:trie.NewStateTrie()}
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
	return db.trie.Root.Value
}

func (db *Database) put(key []byte, value []byte) error {
	err := db.db.Put(key, value, nil)
	if err == nil {
		db.trie.Insert(key, value)
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
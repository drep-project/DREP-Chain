package database

import (
	"github.com/syndtr/goleveldb/leveldb"
	"BlockChainTest/trie"
	"BlockChainTest/config"
	"fmt"
	"BlockChainTest/util/list"
	"bytes"
    "errors"
)

const (
	insOut = iota
	insIn
	modOut
	modIn
	delOut
	delIn
)

type journalEntry struct {
	chainId     config.ChainIdType
	action      int
	key         []byte
	prev        []byte
}

type Transactional interface {
	PutOutState(chainId config.ChainIdType, key []byte, value []byte) error
	PutInState(chainId config.ChainIdType, key []byte, value []byte) error
	Get(key []byte) []byte
	DeleteOutState(chainId config.ChainIdType, key []byte) error
	DeleteInState(chainId config.ChainIdType, key []byte) error
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

func (t *Transaction) PutOutState(chainId config.ChainIdType, key []byte, value []byte) error {
	if t.finished {
		return nil
	}
	prev := t.parent.Get(key)
	if prev == nil {
		t.journal = append(t.journal, &journalEntry{chainId: chainId, action: insOut, key: key})
	} else {
		t.journal = append(t.journal, &journalEntry{chainId: chainId, action: modOut, key: key, prev:prev})
	}
	return t.parent.PutOutState(chainId, key, value)
}

func (t *Transaction) PutInState(chainId config.ChainIdType, key []byte, value []byte) error {
    if t.finished {
        return nil
    }
    prev := t.parent.Get(key)
    if prev == nil {
        t.journal = append(t.journal, &journalEntry{chainId: chainId, action: insIn, key: key})
    } else {
        t.journal = append(t.journal, &journalEntry{chainId: chainId, action: modIn, key: key, prev:prev})
    }
    return t.parent.PutInState(chainId, key, value)
}

func (t *Transaction) Get(key []byte) []byte {
	if t.finished {
		return nil
	}
	return t.parent.Get(key)
}

func (t *Transaction) DeleteOutState(chainId config.ChainIdType, key []byte) error {
	if t.finished {
		return nil
	}
	prev := t.parent.Get(key)
	if prev == nil {
		return errors.New("no such key found in database")
	}
	t.journal = append(t.journal, &journalEntry{chainId:chainId, action: delOut, key:key, prev:prev})
	return t.parent.DeleteOutState(chainId, key)
}

func (t *Transaction) DeleteInState(chainId config.ChainIdType, key []byte) error {
    if t.finished {
        return nil
    }
    prev := t.parent.Get(key)
    if prev == nil {
        return errors.New("no such key found in database")
    }
    t.journal = append(t.journal, &journalEntry{chainId:chainId, action: delIn, key:key, prev:prev})
    return t.parent.DeleteInState(chainId, key)
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
	for i := len(t.journal) - 1; i >= 0; i-- {
		e := t.journal[i]
		switch e.action {
		case insOut:
			t.parent.DeleteOutState(e.chainId, e.key)
        case insIn:
            t.parent.DeleteInState(e.chainId, e.key)
		case modOut:
			t.parent.PutOutState(e.chainId, e.key, e.prev)
        case modIn:
            t.parent.PutInState(e.chainId, e.key, e.prev)
		case delOut:
			t.parent.PutOutState(e.chainId, e.key, e.prev)
        case delIn:
            t.parent.PutInState(e.chainId, e.key, e.prev)
		}
	}
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

func (db *Database) PutOutState(chainId config.ChainIdType, key []byte, value []byte) error {
    if err := db.db.Put(key, value, nil); err != nil {
        fmt.Println("error occurs", err)
        return err
    }
    return nil
}

func (db *Database) PutInState(chainId config.ChainIdType, key []byte, value []byte) error {
	if err := db.db.Put(key, value, nil); err == nil {
		t, exists := db.tries[chainId]
		if !exists {
			t = trie.NewStateTrie()
			db.tries[chainId] = t
		}
		t.Insert(key, value)
		return nil
	} else {
		fmt.Println("error occurs", err)
		return err
	}
}

func (db *Database) Get(key []byte) []byte {
	if ret, err := db.db.Get(key, nil); err == nil {
		return ret
	} else {
		return nil
	}
}

func (db *Database) DeleteOutState(chainId config.ChainIdType, key []byte) error {
    if err := db.db.Delete(key, nil); err != nil {
        fmt.Println("Error occurs.", err)
        return err
    }
    return nil
}

func (db *Database) DeleteInState(chainId config.ChainIdType, key []byte) error {
	if err := db.db.Delete(key, nil); err == nil {
		t, exists := db.tries[chainId]
		if !exists {
			fmt.Println("What the fuck, the trie dose not exist.")
			t = trie.NewStateTrie()
			db.tries[chainId] = t
		}
		t.Delete(key)
		return nil
	} else {
		fmt.Println("Error occurs.", err)
		return err
	}
}

func (db *Database) Commit() {}

func (db *Database) Discard() {}

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
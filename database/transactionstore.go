package database

import (
	"github.com/drep-project/DREP-Chain/common/trie"
	"sync"
)

type TransactionStore struct {
	dirties *sync.Map //The data belongs to storage's cache
	trie    *trie.SecureTrie
}
type SnapShot dirtiesKV
type dirtiesKV struct {
	storageDirties *sync.Map //The data belongs to storage's cache
}

func NewTransactionStore(trie *trie.SecureTrie) *TransactionStore {
	return &TransactionStore{
		dirties: new(sync.Map),
		trie:    trie,
	}
}

func (tDb *TransactionStore) Get(key []byte) ([]byte, error) {
	if val, ok := tDb.dirties.Load(string(key)); ok {
		if val == nil {
			return nil, nil
		}
		return val.([]byte), nil
	}
	val, err := tDb.trie.TryGet(key)
	if err != nil {
		return nil, err
	}
	tDb.dirties.Store(string(key), val)
	return val, nil
}

func (tDb *TransactionStore) Put(key []byte, value []byte) error {
	tDb.dirties.Store(string(key), value)
	return nil
}

func (tDb *TransactionStore) Delete(key []byte) error {
	tDb.dirties.Store(string(key), nil)
	return nil
}

func (tDb *TransactionStore) Flush() {
	tDb.dirties.Range(func(key, value interface{}) bool {
		bk := []byte(key.(string))
		if value != nil {
			val := value.([]byte)
			err := tDb.trie.TryUpdate(bk, val)
			if err != nil {
				log.Error("Flush():", err)
				panic(err)
			}

			tDb.trie.Commit(nil)
			return true
		} else {
			tDb.trie.Delete(bk)
			tDb.trie.Commit(nil)
		}
		tDb.dirties.Delete(key)
		return true
	})
}

func (tDb *TransactionStore) RevertState(snapShot *SnapShot) {
	tDb.dirties = snapShot.storageDirties
}

func (tDb *TransactionStore) CopyState() *SnapShot {
	newDirties := dirtiesKV{}

	newMap := new(sync.Map)
	tDb.dirties.Range(func(key, value interface{}) bool {
		if value == nil {
			newMap.Store(key, value)
		} else {
			switch t := value.(type) {
			case []byte:
				newBytes := make([]byte, len(t))
				copy(newBytes, t)
				newMap.Store(key, newBytes)
			default:
				panic("never run here")
			}
		}
		return true
	})

	newDirties.storageDirties = newMap
	return (*SnapShot)(&newDirties)
}

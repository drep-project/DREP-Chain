package database

import (
	"sync"
)

type TransactionDatabase struct {
	store   IStore
	dirties *sync.Map
}

func NewTransactionDatabase(store IStore) *TransactionDatabase {
	return &TransactionDatabase{
		store:   store,
		dirties: new(sync.Map),
	}
}
func (tDb *TransactionDatabase) Get(key []byte) ([]byte, error) {
	if val, ok := tDb.dirties.Load(string(key)); ok {
		return val.([]byte), nil
	}
	val, err := tDb.store.Get(key)
	if err != nil {
		return nil, err
	}
	tDb.dirties.Store(string(key), val)
	return val, nil
}

func (tDb *TransactionDatabase) Put(key []byte, value []byte) error {
	tDb.dirties.Store(string(key), value)
	return nil
}

func (tDb *TransactionDatabase) Delete(key []byte) error {
	tDb.dirties.Store(string(key), nil)
	return nil
}

func (tDb *TransactionDatabase) NewIterator(key []byte) Iterator {
	panic(ErrKeyUnSpport)
}

func (tDb *TransactionDatabase) Flush() {
	tDb.dirties.Range(func(key, value interface{}) bool {
		bk := []byte(key.(string))
		val := value.([]byte)
		if value != nil {
			err := tDb.store.Put(bk, val)
			if err != nil {
				return false
			}
		} else {
			err := tDb.store.Delete(bk)
			if err != nil {
				return false
			}
		}
		return true
	})
}

func (tDb *TransactionDatabase) Clear() {
	tDb.dirties = new(sync.Map)
}

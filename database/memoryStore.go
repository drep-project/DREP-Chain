package database

import (
	"sync"
)

type MemoryStore struct {
	db *sync.Map
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		db: new(sync.Map),
	}
}

func (mStore *MemoryStore) Get(key []byte) ([]byte, error) {
	if val, ok := mStore.db.Load(string(key)); ok {
		if val == nil {
			return nil, nil
		}
		return val.([]byte), nil
	}
	return nil, ErrKeyNotFound
}

func (mStore *MemoryStore) Put(key []byte, value []byte) error {
	mStore.db.Store(string(key), value)
	return nil
}

func (mStore *MemoryStore) Delete(key []byte) error {
	mStore.db.Delete(string(key))
	return nil
}

func (mStore *MemoryStore) NewIterator(key []byte) Iterator {
	panic(ErrKeyUnSpport)
}

package database

import (
	"errors"
	"sync"
)

type TransactionDatabase struct {
	store IStore
	temp  *sync.Map
}

func NewTransactionDatabase(store IStore) *TransactionDatabase {
	return &TransactionDatabase{
		store: store,
		temp: new (sync.Map),
	}
}
func (tDb *TransactionDatabase) Get(key []byte) ([]byte, error) {
	if val, ok := tDb.temp.Load(string(key)); ok {
		return val.([]byte), nil
	}
	val, err := tDb.store.Get(key)
	if err != nil {
		return nil, err
	}
	tDb.temp.Store(string(key), val)
	return val, nil
}

func (tDb *TransactionDatabase) Put(key []byte, value  []byte)  error {
	 tDb.temp.Store(string(key), value)
	 return nil
}

func (tDb *TransactionDatabase) Delete(key []byte)  error {
	tDb.temp.Store(string(key), nil)
	return nil
}

func (tDb *TransactionDatabase) NewIterator(key []byte) Iterator {
   panic(errors.New("not support iterator"))
}

func (tDb *TransactionDatabase) Flush()  {
	tDb.temp.Range(func(key, value interface{}) bool {
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

func (tDb *TransactionDatabase) Clear()  {
	tDb.temp = new (sync.Map)
}
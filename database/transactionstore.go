package database

import (
	"github.com/drep-project/binary"
	"math/big"
	"strconv"
	"sync"
)

type TransactionStore struct {
	store   IStore
	dirties *sync.Map
}

func NewTransactionStore(store IStore) *TransactionStore {
	return &TransactionStore{
		store:   store,
		dirties: new(sync.Map),
	}
}

func (tDb *TransactionStore) Get(key []byte) ([]byte, error) {
	if val, ok := tDb.dirties.Load(string(key)); ok {
		if val == nil {
			return nil, nil
		}
		return val.([]byte), nil
	}
	val, err := tDb.store.Get(key)
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

func (tDb *TransactionStore) NewIterator(key []byte) Iterator {
	panic(ErrKeyUnSpport)
}

func (tDb *TransactionStore) Flush(needLog bool) {
	tDb.dirties.Range(func(key, value interface{}) bool {
		bk := []byte(key.(string))
		if value != nil {
			val := value.([]byte)
			if needLog {
				err := tDb.PutOpLog(bk, val)
				if err != nil {
					return false
				}
			}
			err := tDb.store.Put(bk, val)
			if err != nil {
				return false
			}
		} else {
			if needLog {
				err := tDb.PutDelLog(bk)
				if err != nil {
					return false
				}
			}
			err := tDb.store.Delete(bk)
			if err != nil {
				return false
			}

		}
		tDb.dirties.Delete(key)
		return true
	})
}

func (tDb *TransactionStore) PutOpLog(key, value []byte) error {
	seqVal, err := tDb.store.Get([]byte(dbOperaterMaxSeqKey))
	if err != nil {
		return err
	}

	var seq = new(big.Int).SetBytes(seqVal).Int64() + 1
	previous, _ := tDb.store.Get(key)
	j := &journal{
		Op:       "put",
		Key:      key,
		Value:    value,
		Previous: previous,
	}
	err = tDb.store.Put(key, value)
	if err != nil {
		return err
	}
	jVal, err := binary.Marshal(j)
	if err != nil {
		return err
	}
	//存储seq-operater kv对
	err = tDb.store.Put([]byte(dbOperaterJournal+strconv.FormatInt(seq, 10)), jVal)
	if err != nil {
		return err
	}
	//记录当前最高的seq
	return tDb.store.Put([]byte(dbOperaterMaxSeqKey), new(big.Int).SetInt64(seq).Bytes())
}

func (tDb *TransactionStore) PutDelLog(key []byte) error {
	seqVal, err := tDb.store.Get([]byte(dbOperaterMaxSeqKey))
	if err != nil {
		return err
	}
	var seq = new(big.Int).SetBytes(seqVal).Int64() + 1
	previous, _ := tDb.store.Get(key)
	j := &journal{
		Op:       "del",
		Key:      key,
		Previous: previous,
	}
	err = tDb.store.Delete(key)
	if err != nil {
		return err
	}
	jVal, err := binary.Marshal(j)
	if err != nil {
		return err
	}
	err = tDb.store.Put([]byte(dbOperaterJournal+strconv.FormatInt(seq, 10)), jVal)
	if err != nil {
		return err
	}
	err = tDb.store.Put([]byte(dbOperaterMaxSeqKey), new(big.Int).SetInt64(seq).Bytes())
	if err != nil {
		return err
	}
	return tDb.store.Delete(key)
}

func (tDb *TransactionStore) Clear() {
	tDb.dirties = new(sync.Map)
}

func (tDb *TransactionStore) RevertState(dirties *sync.Map) {
	tDb.dirties = dirties
}

func (tDb *TransactionStore) CopyState() *sync.Map {
	return copyMap(tDb.dirties)
}

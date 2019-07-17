package database

import (
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/database/drepdb"
	"github.com/drep-project/drep-chain/database/trie"
	"math/big"
	"strconv"
	"sync"
)

type TransactionStore struct {
	diskDB  drepdb.KeyValueStore //本对象内，仅仅作为存储操作日志
	dirties *sync.Map
	trie    *trie.SecureTrie
}

func NewTransactionStore(trie *trie.SecureTrie, diskDB drepdb.KeyValueStore) *TransactionStore {
	return &TransactionStore{
		diskDB:  diskDB,
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

func (tDb *TransactionStore) Flush(needLog bool) {
	tDb.dirties.Range(func(key, value interface{}) bool {
		bk := []byte(key.(string))
		if value != nil {
			val := value.([]byte)
			if needLog {
				err := tDb.putOpLog(bk, val)
				if err != nil {
					return false
				}
			}
			err := tDb.trie.TryUpdate(bk, val)
			if err != nil {
				return false
			}
			tDb.trie.Commit(nil)
		} else {
			if needLog {
				err := tDb.putDelLog(bk)
				if err != nil {
					return false
				}
			}

			tDb.trie.Delete(bk)
			tDb.trie.Commit(nil)
		}
		tDb.dirties.Delete(key)
		return true
	})
}

func (tDb *TransactionStore) putOpLog(key, value []byte) error {
	seqVal, err := tDb.diskDB.Get([]byte(dbOperaterMaxSeqKey))
	if err != nil {
		return err
	}

	var seq = new(big.Int).SetBytes(seqVal).Int64() + 1
	previous, _ := tDb.diskDB.Get(key)
	j := &journal{
		Op:       "put",
		Key:      key,
		Value:    value,
		Previous: previous,
	}
	err = tDb.diskDB.Put(key, value)
	if err != nil {
		return err
	}
	jVal, err := binary.Marshal(j)
	if err != nil {
		return err
	}
	//存储seq-operater kv对
	err = tDb.diskDB.Put([]byte(dbOperaterJournal+strconv.FormatInt(seq, 10)), jVal)
	if err != nil {
		return err
	}
	//记录当前最高的seq
	return tDb.diskDB.Put([]byte(dbOperaterMaxSeqKey), new(big.Int).SetInt64(seq).Bytes())
}

func (tDb *TransactionStore) putDelLog(key []byte) error {
	seqVal, err := tDb.diskDB.Get([]byte(dbOperaterMaxSeqKey))
	if err != nil {
		return err
	}
	var seq = new(big.Int).SetBytes(seqVal).Int64() + 1
	previous, _ := tDb.diskDB.Get(key)
	j := &journal{
		Op:       "del",
		Key:      key,
		Previous: previous,
	}
	tDb.Delete(key)

	jVal, err := binary.Marshal(j)
	if err != nil {
		return err
	}
	err = tDb.diskDB.Put([]byte(dbOperaterJournal+strconv.FormatInt(seq, 10)), jVal)
	if err != nil {
		return err
	}
	err = tDb.diskDB.Put([]byte(dbOperaterMaxSeqKey), new(big.Int).SetInt64(seq).Bytes())
	if err != nil {
		return err
	}

	tDb.diskDB.Delete(key)
	return nil
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

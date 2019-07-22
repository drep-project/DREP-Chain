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
	dirties *dirtiesKV
	trie    *trie.SecureTrie
}

type dirtiesKV struct {
	storageDirties *sync.Map //数据属于storage的缓存
	otherDirties   *sync.Map //数据与stroage没有关系的，其他kv对的缓存
}

func NewTransactionStore(trie *trie.SecureTrie, diskDB drepdb.KeyValueStore) *TransactionStore {
	return &TransactionStore{
		diskDB: diskDB,
		dirties: &dirtiesKV{
			otherDirties:   new(sync.Map),
			storageDirties: new(sync.Map),
		},
		trie: trie,
	}
}

func (tDb *TransactionStore) Get(key []byte) ([]byte, error) {
	if val, ok := tDb.dirties.storageDirties.Load(string(key)); ok {
		if val == nil {
			return nil, nil
		}
		return val.([]byte), nil
	}
	val, err := tDb.trie.TryGet(key)
	if err != nil {
		return nil, err
	}
	tDb.dirties.storageDirties.Store(string(key), val)
	return val, nil
}

func (tDb *TransactionStore) Put(key []byte, value []byte) error {
	tDb.dirties.storageDirties.Store(string(key), value)
	return nil
}

func (tDb *TransactionStore) Delete(key []byte) error {
	tDb.dirties.storageDirties.Store(string(key), nil)
	return nil
}

func (tDb *TransactionStore) Flush(needLog bool) {
	tDb.dirties.storageDirties.Range(func(key, value interface{}) bool {
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
		tDb.dirties.storageDirties.Delete(key)
		return true
	})


	tDb.dirties.otherDirties.Range(func(key, value interface{}) bool {
		bk := []byte(key.(string))
		val := value.([]byte)
		tDb.diskDB.Put(bk, val)
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
	tDb.dirties.storageDirties = new(sync.Map)
	tDb.dirties.otherDirties = new(sync.Map)
}

func (tDb *TransactionStore) RevertState(dirties *dirtiesKV) {
	tDb.dirties = dirties
}

func (tDb *TransactionStore) CopyState() *SnapShot {
	newDirties := dirtiesKV{}

	newMap := new(sync.Map)
	tDb.dirties.storageDirties.Range(func(key, value interface{}) bool {
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

	newMap = new(sync.Map)
	tDb.dirties.otherDirties.Range(func(key, value interface{}) bool {
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

	newDirties.otherDirties = newMap
	return (*SnapShot)(&newDirties)
}

func copyDirMap(m *sync.Map) *sync.Map {
	newMap := new(sync.Map)
	m.Range(func(key, value interface{}) bool {
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
	return newMap
}

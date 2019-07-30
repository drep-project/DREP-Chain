package database

import (
	"sync"

	"github.com/drep-project/drep-chain/database/drepdb"
	"github.com/drep-project/drep-chain/database/trie"
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

//func (tDb *TransactionStore) Delete(key []byte) error {
//	tDb.dirties.storageDirties.Store(string(key), nil)
//	return nil
//}

func (tDb *TransactionStore) Flush() {
	tDb.dirties.storageDirties.Range(func(key, value interface{}) bool {
		bk := []byte(key.(string))
		if value != nil {
			val := value.([]byte)
			err := tDb.trie.TryUpdate(bk, val)
			if err != nil {
				return false
			}
			tDb.trie.Commit(nil)
		} else {
			tDb.trie.Delete(bk)
			tDb.trie.Commit(nil)
		}
		tDb.dirties.storageDirties.Delete(key)
		return true
	})

	// todo 把otherDirties的内容合并到 storageDirties中，便于回滚的状态一致性
	tDb.dirties.otherDirties.Range(func(key, value interface{}) bool {
		bk := []byte(key.(string))
		val := value.([]byte)
		tDb.diskDB.Put(bk, val)
		return true
	})
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

//func copyDirMap(m *sync.Map) *sync.Map {
//	newMap := new(sync.Map)
//	m.Range(func(key, value interface{}) bool {
//		if value == nil {
//			newMap.Store(key, value)
//		} else {
//			switch t := value.(type) {
//			case []byte:
//				newBytes := make([]byte, len(t))
//				copy(newBytes, t)
//				newMap.Store(key, newBytes)
//			default:
//				panic("never run here")
//			}
//		}
//		return true
//	})
//	return newMap
//}

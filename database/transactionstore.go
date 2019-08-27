package database

import (
	"sync"

	"github.com/drep-project/drep-chain/database/drepdb"
	"github.com/drep-project/drep-chain/database/trie"
)

type TransactionStore struct {
	diskDB  drepdb.KeyValueStore //本对象内，仅仅作为存储操作日志
	dirties *sync.Map            //数据属于storage的缓存
	trie    *trie.SecureTrie
}

type dirtiesKV struct {
	storageDirties *sync.Map //数据属于storage的缓存
	//otherDirties   *sync.Map //数据与stroage没有关系的，其他kv对的缓存
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

func (tDb *TransactionStore) Flush() {
	tDb.dirties.Range(func(key, value interface{}) bool {
		bk := []byte(key.(string))
		if value != nil {
			val := value.([]byte)
			err := tDb.trie.TryUpdate(bk, val)
			if err != nil {
				log.Error("Flush():",err)
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

func (tDb *TransactionStore) RevertState(dirties *sync.Map) {
	tDb.dirties = dirties
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

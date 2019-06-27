package database

import (
	"github.com/drep-project/drep-chain/common/fileutil"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"sync"
)

type LdbStore struct {
	db *leveldb.DB
}

func NewLdbStore(dbPath string) (*LdbStore, error) {
	fileutil.EnsureDir(dbPath)
	ldb, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, err
	}
	db := &LdbStore{
		db: ldb,
	}
	return db, nil
}

func (ldbStorte *LdbStore) Get(key []byte) ([]byte, error) {
	return ldbStorte.db.Get(key, nil)
}

func (ldbStorte *LdbStore) Put(key []byte, value []byte) error {
	return ldbStorte.db.Put(key, value, nil)
}

func (ldbStorte *LdbStore) Delete(key []byte) error {
	return ldbStorte.db.Delete(key, nil)
}

func (ldbStorte *LdbStore) NewIterator(key []byte) Iterator {
	return ldbStorte.db.NewIterator(util.BytesPrefix(key), nil)
}

func (ldbStorte *LdbStore) RevertState(dirties *sync.Map) {
	panic(ErrKeyUnSpport)
}

func (ldbStorte *LdbStore) CopyState() *sync.Map {
	panic(ErrKeyUnSpport)
}

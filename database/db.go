package database

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"encoding/hex"
	"sync"
	"BlockChainTest/trie"
)

var db *Database
var once sync.Once

type Database struct {
	Name      string
	LevelDB   *leveldb.DB
	Trie      *trie.StateTrie
}

var databaseName = "local_data"

func NewDatabase() *Database {
	ldb, err := leveldb.OpenFile(databaseName, nil)
	if err != nil {
		panic(err)
	}
	return &Database{
		Name: databaseName,
		LevelDB: ldb,
		Trie: trie.NewStateTrie(),
	}
}

func GetDatabase() *Database {
	once.Do(func() {
		if db == nil {
			db = NewDatabase()
		}
	})
	return db
}

func (db *Database) Delete(key string) error {
	k, err := hex.DecodeString(key)
	if err != nil {
		return err
	}
	return db.LevelDB.Delete(k, nil)
}

func (db *Database) Store(key, value []byte) error {
	return db.LevelDB.Put(key, value, nil)
}

func (db *Database) Load(key []byte) ([]byte, error) {
	return db.LevelDB.Get(key, nil)
}

type Iterator struct {
	Itr iterator.Iterator
}

func (db *Database) NewIterator() *Iterator {
	return &Iterator{db.LevelDB.NewIterator(nil, nil)}
}

func (itr *Iterator) Next() bool {
	return itr.Itr.Next()
}

func (itr *Iterator) Key() []byte {
	return itr.Itr.Key()
}

func (itr *Iterator) Value() []byte {
	return itr.Itr.Value()
}

func (itr *Iterator) Release() {
	itr.Itr.Release()
}

func GetStateRoot() []byte {
	db := GetDatabase()
	return db.Trie.Root.Value
}
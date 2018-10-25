package database

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"encoding/hex"
	"sync"
	"strconv"
)

var db *Database
var once sync.Once

type Database struct {
	LevelDB *leveldb.DB
	Name    string
}

var databaseName = "local_data"

func NewDatabase() *Database {
	ldb, err := leveldb.OpenFile(databaseName, nil)
	defer ldb.Close()
	if err != nil {
		panic(err)
	}
	return &Database{ldb, databaseName}
}

func GetDatabase() *Database {
	once.Do(func() {
		if db == nil {
			db = NewDatabase()
		}
	})
	return db
}

func (db *Database) Get(key string) (DBElem, error) {
	k, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}
	b, err := db.LevelDB.Get(k, nil)
	if err != nil {
		return nil, err
	}
	return unmarshal(b)
}

func (db *Database) Put(elem DBElem) (string, []byte, error) {
	key := elem.DBKey()
	k, err := hex.DecodeString(key)
	if err != nil {
		return "", nil, err
	}
	b, err := marshal(elem)
	if err != nil {
		return "", nil, err
	}
	return key, b, db.LevelDB.Put(k, b, nil)
}

func (db *Database) PutInt(key string, value int) {
	db.LevelDB.Put([]byte(key), []byte(strconv.Itoa(value)), nil)
}

func (db *Database) GetInt(key string) (int, error) {
	if value, err := db.LevelDB.Get([]byte(key), nil); err == nil {
		if r, err := strconv.Atoi(string(value)); err == nil {
			return r, nil
		} else {
			return 0, err
		}
	} else {
		return 0, err
	}
}

func (db *Database) Delete(key string) error {
	k, err := hex.DecodeString(key)
	if err != nil {
		return err
	}
	return db.LevelDB.Delete(k, nil)
}

func (db *Database) Open() {
	var err error
	db.LevelDB, err = leveldb.OpenFile(db.Name, nil)
	if err != nil {
		panic(err)
	}
	return
}

func (db *Database) Close() {
	db.LevelDB.Close()
}

func (db *Database) Clear()  {

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

func (itr *Iterator) Key() string {
	return string(itr.Itr.Key())
}

func (itr *Iterator) Value() []byte {
	return itr.Itr.Value()
}

func (itr *Iterator) Elem() (DBElem, error) {
	return unmarshal(itr.Itr.Value())
}

func (itr *Iterator) Release() {
	itr.Itr.Release()
}
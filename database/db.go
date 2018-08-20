package database

import (
    "github.com/syndtr/goleveldb/leveldb"
    "BlockChainTest/network"
    "github.com/syndtr/goleveldb/leveldb/iterator"
)

type Database struct {
    LevelDB *leveldb.DB
    FilePath string
}

func NewDatabase(filePath string) (*Database, error) {
   ldb, err := leveldb.OpenFile(filePath, nil)
   return &Database{ldb, filePath}, err
}

func (db *Database) Get(key string) (interface{}, error) {
    b, err := db.LevelDB.Get([]byte(key), nil)
    if err != nil {
        return nil, err
    }
    value, err := network.Deserialize(b)
    if err != nil {
        return nil, err
    }
    return value, nil
}

func (db *Database) Put(key string, value interface{}) error {
    b, err := network.Serialize(value)
    if err != nil {
        return err
    }
    return db.LevelDB.Put([]byte(key), b, nil)
}

func (db *Database) Delete(key string) error {
    return db.LevelDB.Delete([]byte(key), nil)
}

func (db *Database) Close() error {
    return db.LevelDB.Close()
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
    return string(itr.Key())
}

func (itr *Iterator) Value() interface{} {
    value, _ := network.Deserialize(itr.Itr.Value())
    return value
}
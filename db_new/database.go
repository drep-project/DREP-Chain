package db_new

import (
    "github.com/syndtr/goleveldb/leveldb"
    "github.com/pkg/errors"
)

type Database struct {
    db *leveldb.DB
    temp map[string] []byte
    runningChain string
}


func (db *Database) put(key []byte, value []byte, temporary bool) error {
    if !temporary {
        return db.db.Put(key, value, nil)
    }
    if db.temp == nil {
        return errors.New("put error: no temp db opened")
    }
    db.temp[bytes2Hex(key)] = value
    return nil
}

func (db *Database) get(key []byte, temporary bool) ([]byte, error) {
    if !temporary {
        return db.db.Get(key, nil)
    }
    if db.temp == nil {
        db.temp = make(map[string] []byte)
    }
    hk := bytes2Hex(key)
    value, ok := db.temp[hk]
    if !ok {
        var err error
        value, err = db.db.Get(key, nil)
        if err != nil {
            return nil, err
        }
        db.temp[hk] = value
    }
    return value, nil
}

func (db *Database) delete(key []byte, temporary bool) error {
    if !temporary {
        return db.db.Delete(key, nil)
    }
    if db.temp == nil {
        return errors.New("delete error: no temp db opened")
    }
    db.temp[bytes2Hex(key)] = nil
    return nil
}

func (db *Database) Commit() {
    if db.temp == nil {
        return
    }
    for key, value := range db.temp {
        bk := hex2Bytes(key)
        if value != nil {
            db.put(bk, value, false)
        } else {
            db.delete(bk, false)
        }
    }
    db.temp = nil
}

func (db *Database) Discard() {
    db.temp = nil
}
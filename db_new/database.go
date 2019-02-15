package db_new

import (
    "github.com/syndtr/goleveldb/leveldb"
    "github.com/pkg/errors"
    "encoding/json"
)

type Database struct {
    db           *leveldb.DB
    temp         map[string] []byte
    states       map[string] *State
    runningChain string
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

func (db *Database) getState(key []byte) (*State, error) {
    var state *State
    if db.states == nil {
        db.states = make(map[string] *State)
    }
    hk := bytes2Hex(key)
    state, ok := db.states[hk]
    if ok {
        return state, nil
    }
    b, err := db.get(key, true)
    if err != nil {
        return nil, err
    }
    state = &State{}
    err = json.Unmarshal(b, state)
    if err != nil {
        return nil, err
    }
    db.states[hk] = state
    return state, nil
}

func (db *Database) putState(key []byte, state *State) error {
    if db.states == nil {
        db.states = make(map[string] *State)
    }
    b, err := json.Marshal(state)
    if err != nil {
        return err
    }
    err = db.put(key, b, true)
    if err != nil {
        return err
    }
    db.states[bytes2Hex(key)] = state
    return nil
}

func (db *Database) deleteState(key []byte) error {
    if db.states == nil {
        db.states = make(map[string] *State)
    }
    err := db.delete(key, true)
    if err != nil {
        return err
    }
    db.states[bytes2Hex(key)] = nil
    return nil
}
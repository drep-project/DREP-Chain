package db

import "github.com/syndtr/goleveldb/leveldb"

const maxDbWaitingSize = 100

type Database struct {
    cache       *LRU
    db          *leveldb.DB
    waiting     []*cacheEntry
    waitingSize int
    channel     chan *cacheEntry
    closeCh     chan bool
}

func NewDatabase(file string) (*Database, error) {
    if db, err := leveldb.OpenFile(file, nil); err != nil {
        return nil, err
    } else {
        channel := make(chan *cacheEntry, 100) // TODO
        closeCh := make(chan bool)
        lru, _ := NewLRU(10000, nil)
        db := &Database{
            cache:          lru,
            db:             db,
            waiting:        make([]*cacheEntry, maxDbWaitingSize),
            waitingSize:    0,
            channel:        channel,
            closeCh:        closeCh,
        }
        go func() {
            for {
                select {
                case c := <- channel:
                    db.waiting[db.waitingSize] = c
                    db.waitingSize++
                    if db.waitingSize >= maxDbWaitingSize {
                        db.flush()
                    }
                case <- closeCh:
                    db.flush()
                    db.db.Close()
                    break
                }
            }
        }()
        return db, nil
    }
}

func (db *Database) flush() {
    defer func() {
        db.waitingSize = 0
    }()
    t, err := db.db.OpenTransaction()
    if err != nil {
        return
    }
    for i := 0; i < db.waitingSize; i++ {
        w := db.waiting[i]
        if w.val == nil {
            t.Delete(w.key, nil)
        } else {
            t.Put(w.key, w.val, nil)
        }
    }
    t.Commit()
}

func (db *Database) close() {
    db.closeCh <- true
}


func (db *Database) Put(key []byte, value []byte) error {
    db.cache.Add(string(key), value)
    db.channel <- &cacheEntry{key: key, val: value}
    return nil
}

func (db *Database) Get(key []byte) ([]byte, error) {
    if v, exists := db.cache.Get(string(key)); exists {
        return v.([]byte), nil
    } else if v, err := db.db.Get(key, nil); err == nil {
        return v, nil
    } else {
        return nil, err
    }
}

func (db *Database) Delete(key []byte) error {
    db.cache.Remove(string(key))
    db.channel <- &cacheEntry{key: key}
    return nil
}

func (db *Database) BeginTransaction() Transactional {
    return nil
}
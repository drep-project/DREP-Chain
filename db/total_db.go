package db

type StateDb struct {
    db      *Database
}

func NewTotalDb(file string) *StateDb {
    if db, err := NewDatabase(file); err == nil {
        return &StateDb{
            db: db,
        }
    } else {
        return nil
    }
}

func (db *StateDb) Put(key []byte, value []byte) error {
    err := db.db.Put(key, value)
    return err
}

func (db *StateDb) Get(key []byte) ([]byte, error) {
    return db.db.Get(key)
}

func (db *StateDb) Delete(key []byte) error {
    err := db.db.Delete(key)
    return err
}

func (db *StateDb) BeginTransaction() Transactional {
    return NewTransaction(db, nil)
}

type StateTrie 
package database

import (
    "math/big"
    "encoding/json"
    "strconv"
)

type Journal struct {
    op       string
    key      []byte
    value    []byte
    previous []byte
}

func (db *Database) getJournalLength() int64 {
    var length int64 = -1
    lengthVal, err := db.db.Get([]byte("journal_length"), nil)
    if err == nil {
        length = new(big.Int).SetBytes(lengthVal).Int64()
    }
    return length
}

func (db *Database) addJournal(op string, key, value []byte) error {
    length := db.getJournalLength() + 1
    previous, _ := db.db.Get(key, nil)
    j := &Journal{
        op:       op,
        key:      key,
        value:    value,
        previous: previous,
    }
    jVal, err := json.Marshal(j)
    if err != nil {
        return err
    }
    return db.db.Put([]byte("journal_" + strconv.FormatInt(length, 10)), jVal, nil)
}

func (db *Database) getJournal(index int64) (*Journal, error) {
    jVal, err := db.db.Get([]byte("journal_" + strconv.FormatInt(index, 10)), nil)
    if err != nil {
        return nil, err
    }
    j := &Journal{}
    err = json.Unmarshal(jVal, j)
    return j, err
}

func (db *Database) removeJournal(index int64) error {
    return db.db.Delete([]byte("journal_" + strconv.FormatInt(index, 10)), nil)
}

func (db *Database) revertJournal(index int64) error {
    j, err := db.getJournal(index)
    if err != nil {
        return err
    }
    if j.op == "put" {
        if j.previous == nil {
            return db.db.Delete(j.key, nil)
        } else {
            return db.db.Put(j.key, j.previous, nil)
        }
    }
    if j.op == "del" {
        return db.db.Put(j.key, j.previous, nil)
    }
    return nil
}

func (db *Database) rollback2Index(index int64) error {
    length := db.getJournalLength()
    for i := length; i > index; i-- {
        err := db.revertJournal(i)
        if err != nil {
            return err
        }
        err = db.removeJournal(i)
        if err != nil {
            return err

        }
    }
    return nil
}

func (db *Database) BlockHeight2JournalIndex(height int64) error {
    lengthVal, err := db.db.Get([]byte("journal_lenght"), nil)
    if err != nil {
        return err
    }
    return db.db.Put([]byte("index_of_height_" + strconv.FormatInt(height, 10)), lengthVal, nil)
}

func (db *Database) Rollback2BlockHeight(height int64) error {
    indexVal, err := db.db.Get([]byte("index_of_height_" + strconv.FormatInt(height, 10)), nil)
    if err != nil {
        return err
    }
    return db.rollback2Index(new(big.Int).SetBytes(indexVal).Int64())
}

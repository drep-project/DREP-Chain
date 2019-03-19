package database

import (
    "testing"
    "fmt"
)

var db0 *Database

func init() {
    db0, _ = NewDatabase("tdb")
}

func TestAddJournal(t *testing.T) {
    db0.addJournal("put", []byte{1}, []byte{1})
    db0.addJournal("put", []byte{2}, []byte{2})
    db0.addJournal("put", []byte{3}, []byte{3})
    fmt.Println("length: ", db0.getJournalLength())
}

func TestGetJournal(t *testing.T) {
    fmt.Println(db0.getJournal(1))
    fmt.Println(db0.getJournal(2))
}

func TestRemoveJournal(t *testing.T) {
    db0.removeJournal(1)
    fmt.Println("length: ", db0.getJournalLength())
    db0.removeJournal(0)
    fmt.Println("length: ", db0.getJournalLength())
    fmt.Println(db0.getJournal(1))
}

func TestRollback2Index(t *testing.T) {
    db0.addJournal("put", []byte{11}, []byte{11})
    db0.addJournal("put", []byte{22}, []byte{22})
    db0.addJournal("put", []byte{33}, []byte{33})
    db0.rollback2Index(2)
    fmt.Println(db0.getJournalLength())
    fmt.Println(db0.getJournal(4))
}

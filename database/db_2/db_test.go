package db_2

import (
    "testing"
    "github.com/syndtr/goleveldb/leveldb"
    "fmt"
)

func TestDb1(t *testing.T)  {
    db, err := leveldb.OpenFile("lll", nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer db.Close()
    err = db.Put([]byte("a"), []byte("8"), nil)

    s, _ := db.GetSnapshot()
    err = db.Put([]byte("a"), []byte("9"), nil)
    d, _ := db.Get([]byte("a"), nil)
    fmt.Println(string(d))
    d2, _ := s.Get([]byte("a"), nil)
    fmt.Println(string(d2))
}

func TestDb2(t *testing.T)  {
    db, _ := leveldb.OpenFile("lll", nil)
    db.Put([]byte("a"), []byte("9"), nil)
    tran, _ := db.OpenTransaction()
    defer db.Close()
    tran.Put([]byte("a"), []byte("80"), nil)
    //db.Put([]byte("a"), []byte("9"), nil) // cannot
    s, _ := db.GetSnapshot()
    d, _ := db.Get([]byte("a"), nil)
    fmt.Println(string(d))
    //db.Put([]byte("a"), []byte("100"), nil)
    d2, _ := s.Get([]byte("a"), nil)
    fmt.Println(string(d2))
    tran.Commit()
    d2, _ = db.Get([]byte("a"), nil)
    fmt.Println(string(d2))
}

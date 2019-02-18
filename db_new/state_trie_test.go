package db_new

import (
    "testing"
    "fmt"
)

func TestDBInit(t *testing.T) {
    db = NewDatabase()
    key := []byte("a")
    value := []byte("b")
    db.put(key, value, false)

    v1, err := db.get(key, false)
    fmt.Println("v1:     ", v1)
    fmt.Println("err1:   ", err)
    fmt.Println()

    v2, err := db.get(key, true)
    fmt.Println("v2:     ", v2)
    fmt.Println("err2:   ", err)
    fmt.Println()

    v3, err := db.get(key, true)
    fmt.Println("v3:     ", v3)
    fmt.Println("err2:   ", err)
    fmt.Println()

    fmt.Println("temp:   ", db.temp)
    fmt.Println("states: ", db.states)
    fmt.Println()
}

func TestStateTrie(t *testing.T) {
    db = NewDatabase()
    fmt.Println("root key: ", bytes2Hex(db.rootKey))
    fmt.Println("root:     ", db.root)
    fmt.Println()

    //originKey := mycrypto.Hash256([]byte("a"))
    //originValue := mycrypto.Hash256([]byte("b"))
    //seq := bytes2Hex(originKey)
    //state, err := insert(seq, db.rootKey, originValue)
    //fmt.Println("seq:      ", seq)
    //fmt.Println("value:    ", originValue)
    //fmt.Println("state:    ", state)
    //fmt.Println("err:      ", err)

    insert("a", db.rootKey, []byte{1})
    insert("abb", db.rootKey, []byte{2})
    insert("231", db.rootKey, []byte{3})
    insert("232", db.rootKey, []byte{4})
    //insert("23456", db.rootKey, []byte{4})
    //insert("2ba", db.rootKey, []byte{5})

    fmt.Println()
    fmt.Println("##########################################################################################################")
    fmt.Println()

    search(db.rootKey, "", 0)
}
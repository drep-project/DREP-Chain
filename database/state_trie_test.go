package database
//
//import (
//    "testing"
//    "fmt"
//)
//
//func TestDBInit(t *testing.T) {
//    db = NewDatabase()
//    key := []byte("a")
//    value := []byte("b")
//    db.put(key, value, false)
//
//    v1, err := db.get(key, false)
//    fmt.Println("v1:     ", v1)
//    fmt.Println("err1:   ", err)
//    fmt.Println()
//
//    v2, err := db.get(key, true)
//    fmt.Println("v2:     ", v2)
//    fmt.Println("err2:   ", err)
//    fmt.Println()
//
//    v3, err := db.get(key, true)
//    fmt.Println("v3:     ", v3)
//    fmt.Println("err2:   ", err)
//    fmt.Println()
//
//    fmt.Println("temp:   ", db.temp)
//    fmt.Println("states: ", db.states)
//    fmt.Println()
//}
//
//func TestStateTrie(t *testing.T) {
//    db = NewDatabase()
//    fmt.Println("root Key: ", bytes2Hex(db.root))
//    fmt.Println("root:     ", db.root)
//    fmt.Println()
//
//    //originKey := mycrypto.Hash256([]byte("a"))
//    //originValue := mycrypto.Hash256([]byte("b"))
//    //seq := bytes2Hex(originKey)
//    //state, err := insert(seq, db.root, originValue)
//    //fmt.Println("seq:      ", seq)
//    //fmt.Println("Value:    ", originValue)
//    //fmt.Println("state:    ", state)
//    //fmt.Println("err:      ", err)
//
//    insert("a", db.root, []byte{1})
//    insert("abb", db.root, []byte{2})
//    insert("231", db.root, []byte{3})
//    insert("232", db.root, []byte{4})
//    insert("23456", db.root, []byte{4})
//    insert("2ba", db.root, []byte{5})
//    insert("2bbbbb56", db.root, []byte{6})
//    insert("23147", db.root, []byte{7})
//
//    fmt.Println()
//    fmt.Println("##########################################################################################################")
//    fmt.Println()
//
//    search(db.root, "", 0)
//
//    del(db.root, "abb")
//    del(db.root, "23147")
//
//    fmt.Println()
//    fmt.Println("***********************************************************************************************************")
//    fmt.Println()
//
//    search(db.root, "", 0)
//}
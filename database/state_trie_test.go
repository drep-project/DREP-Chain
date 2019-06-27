package database

//
//import (
//    "testing"
//    "fmt"
//)
//
//func TestDBInit(t *testing.T) {
//    store = NewDatabase()
//    key := []byte("a")
//    value := []byte("b")
//    store.put(key, value, false)
//
//    v1, err := store.get(key, false)
//    fmt.Println("v1:     ", v1)
//    fmt.Println("err1:   ", err)
//    fmt.Println()
//
//    v2, err := store.get(key, true)
//    fmt.Println("v2:     ", v2)
//    fmt.Println("err2:   ", err)
//    fmt.Println()
//
//    v3, err := store.get(key, true)
//    fmt.Println("v3:     ", v3)
//    fmt.Println("err2:   ", err)
//    fmt.Println()
//
//    fmt.Println("dirties:   ", store.dirties)
//    fmt.Println("states: ", store.states)
//    fmt.Println()
//}
//
//func TestStateTrie(t *testing.T) {
//    store = NewDatabase()
//    fmt.Println("root Key: ", bytes2Hex(store.root))
//    fmt.Println("root:     ", store.root)
//    fmt.Println()
//
//    //originKey := mycrypto.Hash256([]byte("a"))
//    //originValue := mycrypto.Hash256([]byte("b"))
//    //seq := bytes2Hex(originKey)
//    //state, err := insert(seq, store.root, originValue)
//    //fmt.Println("seq:      ", seq)
//    //fmt.Println("Value:    ", originValue)
//    //fmt.Println("state:    ", state)
//    //fmt.Println("err:      ", err)
//
//    insert("a", store.root, []byte{1})
//    insert("abb", store.root, []byte{2})
//    insert("231", store.root, []byte{3})
//    insert("232", store.root, []byte{4})
//    insert("23456", store.root, []byte{4})
//    insert("2ba", store.root, []byte{5})
//    insert("2bbbbb56", store.root, []byte{6})
//    insert("23147", store.root, []byte{7})
//
//    fmt.Println()
//    fmt.Println("##########################################################################################################")
//    fmt.Println()
//
//    search(store.root, "", 0)
//
//    del(store.root, "abb")
//    del(store.root, "23147")
//
//    fmt.Println()
//    fmt.Println("***********************************************************************************************************")
//    fmt.Println()
//
//    search(store.root, "", 0)
//}

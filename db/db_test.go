package db

import (
    "testing"
    "fmt"
)

func TestDatabase_BeginTransaction(t *testing.T) {
    b := []byte{1, 2}
    fmt.Println(string(b))
    m := make(map[string][]byte)
    m[string(b)] = nil
    fmt.Println(len(m))
    for k,v := range m {
        fmt.Println("k=", k, "b = ", []byte(k),"v =", v , v ==nil)
    }
    delete(m, string([]byte{1, 2}))
    fmt.Println(len(m))
    for k,v := range m {
        fmt.Println("k=", k, "b = ", []byte(k),"v =", v , v ==nil)
    }
    fmt.Println("J")
    fmt.Println(m[string(b)] == nil)
    m["h"] = nil
    fmt.Println(len(m))
    for k,v := range m {
        fmt.Println("k=", k, "b = ", []byte(k),"v =", v , v ==nil)
    }
}

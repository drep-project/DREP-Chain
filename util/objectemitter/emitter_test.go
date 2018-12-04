package objectemitter

import (
    "testing"
    "time"
    "fmt"
)

func TestObjectEmitter(t *testing.T) {
    e := New(2, 1 * time.Second, func(objects []interface{}) {
        fmt.Println(time.Now(), objects)
    })
    e.Start()
    fmt.Println(time.Now())
    s := 0
    for {
        time.Sleep(700 * time.Millisecond)
        e.Push(s)
        s++
        e.Push(s)
        s++
        time.Sleep(500 * time.Millisecond)
        e.Push(s)
        s++

    }
}

package messagepool

import (
    "testing"
    "time"
    "fmt"
)

func TestMessagePool(t *testing.T) {
    p := NewMessagePool()
    p.Push(34)
    go func() {
        time.Sleep(1 * time.Second)
        p.Push(3)
        p.Push(4)
        time.Sleep(1 * time.Second)
        p.Push(3)
    }()
    fmt.Println(p.Obtain(func(i interface{}) bool {
        if j, ok := i.(int); ok {
            return j == 3
        } else {
            return false
        }
    }, 3 * time.Second))
    fmt.Println(p.Obtain(func(i interface{}) bool {
        if j, ok := i.(int); ok {
            return j == 3
        } else {
            return false
        }
    }, 3 * time.Second))
    fmt.Println(p.Obtain(func(i interface{}) bool {
        if j, ok := i.(int); ok {
            return j == 4
        } else {
            return false
        }
    }, 3 * time.Second))
    fmt.Println(p.Obtain(func(i interface{}) bool {
        if j, ok := i.(int); ok {
            return j == 4
        } else {
            return false
        }
    }, 3 * time.Second))
}
package concurrent

import (
    "testing"
    "time"
    "fmt"
)

func TestExecuteTimeoutTask(t *testing.T) {
    fmt.Println(time.Now())
    fmt.Println(ExecuteTimeoutTask(func() interface{} {
        fmt.Println("OK1")
        time.Sleep(2 * time.Second)
        return nil
    }, 3 * time.Second))
    fmt.Println(time.Now())
    fmt.Println(ExecuteTimeoutTask(func() interface{} {
        fmt.Println("OK2")
        time.Sleep(2 * time.Second)
        return nil
    }, 1 * time.Second))
    fmt.Println(time.Now())
}

func TestNewCountDownLatch(t *testing.T) {
    l := NewCountDownLatch(3)
    fmt.Println(time.Now())
    //go func() {l.Done()}()
    //go func() {l.Done()}()
    //go func() {l.Done()}()
    l.Cancel()
    l.Wait(3 * time.Second)
    fmt.Println(time.Now())
}
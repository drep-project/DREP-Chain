package concurrent

import (
    "time"
    "sync"
    "fmt"
)

type CountDownLatch interface {
    Cancel()
    WaitTimeout(duration time.Duration)
    Wait()
    Done()
}

type countDownLatch struct {
    lock sync.Mutex
    cond *sync.Cond
    ch chan struct{}
    num int
    canceled bool
}

func (l *countDownLatch) Cancel() {
    l.lock.Lock()
    defer l.lock.Unlock()
    if !l.canceled {
        l.canceled = true
        l.ch <- struct{}{}
    } else {
        fmt.Errorf("error: already canceled")
    }
}

func (l *countDownLatch) WaitTimeout(duration time.Duration) {
    select {
    case <- l.ch: case <-time.After(duration):
    }
}

func (l *countDownLatch) Wait() {
    select {
    case <- l.ch:
    }
}

func (l *countDownLatch) Done() {
    l.lock.Lock()
    defer l.lock.Unlock()
    if l.canceled {
        fmt.Errorf("error: already canceled. Cannot done")
        return
    }
    if l.num == 0 {
        fmt.Errorf("error: already done")
        return
    }
    l.num--
    if l.num == 0 {
        l.ch <- struct{}{}
    }
}

func NewCountDownLatch(num int) CountDownLatch {
    c := &countDownLatch{num:num, canceled:false}
    c.cond = sync.NewCond(&c.lock)
    c.ch = make(chan struct{}, 1)// 1 is necessary
    return c
}
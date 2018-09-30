package concurrent

import (
    "time"
    "sync"
)

type CountDownLatch interface {
    Cancel()
    Wait(duration time.Duration)
    Done()
}

type countDownLatch struct {
    lock sync.Mutex
    cond *sync.Cond
    ch chan struct{}
    num int
    canceled bool
}

func (l *countDownLatch) Cancel()  {
    l.lock.Lock()
    defer l.lock.Unlock()
    if !l.canceled {
        l.canceled = true
        l.ch <- struct{}{}
    }
}

func (l *countDownLatch) Wait(duration time.Duration)  {
    select {
    case <- l.ch: case <-time.After(duration):
    }
}

func (l *countDownLatch) Done()  {
    l.lock.Lock()
    defer l.lock.Unlock()
    if l.canceled {
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
    c.ch = make(chan struct{})
    return c
}
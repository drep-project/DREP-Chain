package messagepool

import (
    "sync"
    "time"
)

type MessagePool struct {
    lock sync.Mutex
    cond *sync.Cond
    messages []interface{}
}

//type waitingElement struct {
//    id int64
//    cp func(interface{})bool
//    ch chan interface{}
//}

func NewMessagePool() *MessagePool {
    p := &MessagePool{}
    p.cond = sync.NewCond(&p.lock)
    return p
}

func (p *MessagePool) Obtain(cp func(interface{})bool, duration time.Duration) interface{} {
    p.lock.Lock()
    defer p.lock.Unlock()
    timeout := false
    finish := false
    go func() {
        time.Sleep(duration)
        p.lock.Lock()
        defer p.lock.Unlock()
        if !finish {
            timeout = true
            p.cond.Broadcast()
        }
    }()
    for !timeout {
        for i, m := range p.messages {
            if cp(m) {
                finish = true
                p.messages = append(p.messages[:i], p.messages[i + 1:]...)
                return m
            }
        }
        p.cond.Wait()
    }
    return nil
}

func (p *MessagePool) Push(msg interface{})  {
    p.lock.Lock()
    defer p.lock.Unlock()
    p.messages = append(p.messages, msg)
    p.cond.Broadcast()
}
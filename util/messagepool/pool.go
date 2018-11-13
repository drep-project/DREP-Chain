package messagepool

import (
    "sync"
    "time"
    "BlockChainTest/util/list"
)

type MessagePool struct {
    lock sync.Mutex
    cond *sync.Cond
    messages *list.LinkedList
}

func NewMessagePool() *MessagePool {
    p := &MessagePool{}
    p.cond = sync.NewCond(&p.lock)
    p.messages = list.NewLinkedList()
    return p
}

func (p *MessagePool) Obtain(num int, cp func(interface{})bool, duration time.Duration) []interface{} {
    if num == 0 {
        panic("num should not be 0")
    }
    p.lock.Lock()
    defer p.lock.Unlock()
    timeout := false
    r := make([]interface{}, 0)
    count := 0
    go func() {
        time.Sleep(duration)
        p.lock.Lock()
        defer p.lock.Unlock()
        if count < num {
            timeout = true
            p.cond.Broadcast()
        }
    }()
    for !timeout {
        for it := p.messages.Iterator(); it.HasNext(); {
            m := it.Next()
            if cp(m) {
                r = append(r, m)
                count++
                it.Remove()
                if count == num {
                    return r
                }
            }
        }
        p.cond.Wait()
    }
    return r
}

func (p *MessagePool) ObtainOne(cp func(interface{})bool, duration time.Duration) interface{} {
    if r := p.Obtain(1, cp, duration); len(r) == 1 {
        return r[0]
    } else if len(r) == 0 {
        return nil
    } else {
        panic("what the fuck")
    }

}

func (p *MessagePool) Push(msg interface{})  {
    p.lock.Lock()
    defer p.lock.Unlock()
    p.messages.Add(msg)
    p.cond.Broadcast()
}

func (p *MessagePool) Contains(cp func(interface{})bool) bool {
    p.lock.Lock()
    defer p.lock.Unlock()
    for it := p.messages.Iterator(); it.HasNext(); {
        m := it.Next()
        if cp(m) {
            return true
        }
    }
    return false
}

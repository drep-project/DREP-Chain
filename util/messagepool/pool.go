package messagepool

import (
    "sync"
    "time"
    "BlockChainTest/util/list"
    "fmt"
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
        fmt.Println("ppppppp1")
        if count < num {
            fmt.Println("ppppppp2")
            timeout = true
            p.cond.Broadcast()
        }
        fmt.Println("ppppppp3")
    }()
    fmt.Println("ppppppp4")
    for !timeout {
        fmt.Println("ppppppp5")
        for it := p.messages.Iterator(); it.HasNext(); {
            fmt.Println("ppppppp6")
            m := it.Next()
            if cp(m) {
                fmt.Println("ppppppp7", count, num)
                r = append(r, m)
                count++
                it.Remove()
                if count == num {
                    fmt.Println("ppppppp8")
                    return r
                }
            }
            fmt.Println("ppppppp9")
        }
        p.cond.Wait()
        fmt.Println("ppppppp10")
    }
    fmt.Println("ppppppp11")
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
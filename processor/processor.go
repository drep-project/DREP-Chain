package processor

import (
    "sync"
    "fmt"
)

type processor interface {
    process(interface{})
}

type message struct {
    t int
    msg interface{}
}

var (
    once sync.Once
    singleton *Processor
)

type Processor struct {
    processors map[int]processor
    channel chan *message
}

func (p *Processor) init()  {
    p.channel = make(chan *message)
    p.processors = make(map[int]processor)
    //p.processors[common.MSG_BLOCK] = &confirmedBlockProcessor{}
    //p.processors[common.MSG_TRANSACTION] = &transactionProcessor{}
    

}

func GetInstance() *Processor {
    once.Do(func() {
        singleton = new(Processor)
        singleton.init()
    })
    return singleton
}

func (p *Processor) Start() {
    go func() {
        for {
            if msg, ok := <-p.channel; ok {
                p.dispatch(msg)
            }
        }
    }()
}

func (p *Processor) Process(t int, msg interface{}) {
    p.channel <- &message{t, msg}
}

func (p *Processor) dispatch(msg *message) {
    // TODO something
    if processor := p.processors[msg.t]; processor != nil {
        processor.process(msg.msg)
    } else {
        fmt.Errorf("invalid message %v", msg)
    }
}

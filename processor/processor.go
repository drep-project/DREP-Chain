package processor

import (
    "sync"
    "fmt"
)

type processor interface {
    process(interface{})
}

var (
    once sync.Once
    singleton *Processor
)

type Processor struct {
    processors map[int]processor
    channel chan *common.Message
}

func (p *Processor) init(channel chan *common.Message)  {
    p.channel = channel
    p.processors = make(map[int]processor)
    p.processors[common.MSG_BLOCK] = &confirmedBlockProcessor{}
    p.processors[common.MSG_TRANSACTION] = &transactionProcessor{}

}

func GetInstance(channel chan *common.Message) *Processor {
    once.Do(func() {
        singleton = new(Processor)
        singleton.init(channel)
    })
    return singleton
}

func (p *Processor) Start() {
    go func() {
        for {
            if msg, ok := <-p.channel; ok {
                p.dispatch(*msg)
            }
        }
    }()
}

func (p *Processor) dispatch(msg common.Message) {
    // TODO something
    if processor := p.processors[msg.Type]; processor != nil {
        processor.process(msg.Body)
    } else {
        fmt.Errorf("invalid message %v", msg)
    }
}

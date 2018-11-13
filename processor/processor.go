package processor

import (
    "sync"
    "fmt"
    "BlockChainTest/bean"
    "BlockChainTest/network"
)

type processor interface {
    process(peer *network.Peer, msg interface{})
}

type message struct {
    peer *network.Peer
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
    p.processors[bean.MsgTypeSetUp] = &SetUpProcessor{}
    p.processors[bean.MsgTypeChallenge] = &ChallengeProcessor{}
    p.processors[bean.MsgTypeCommitment] = &CommitProcessor{}
    p.processors[bean.MsgTypeResponse] = &ResponseProcessor{}
    p.processors[bean.MsgTypeBlock] = &BlockProcessor{p}
    p.processors[bean.MsgTypeTransaction] = &transactionProcessor{}
    p.processors[bean.MsgTypeNewPeer] = &NewComerProcessor{}
    p.processors[bean.MsgTypePeerList] = &PeersProcessor{}
    p.processors[bean.MsgTypeBlockReq] = &BlockReqProcessor{}
    p.processors[bean.MsgTypeBlockResp] = &BlockRespProcessor{}
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

func (p *Processor) Process(peer *network.Peer, t int, msg interface{}) {
    p.channel <- &message{peer:peer, t: t, msg:msg}
}

func (p *Processor) dispatch(msg *message) {
    if msg.t == bean.MsgTypeTransaction {
        fmt.Println("Receive transaction")
    }
    if processor := p.processors[msg.t]; processor != nil {
        processor.process(msg.peer, msg.msg)
    } else {
        fmt.Errorf("invalid message %v", msg)
    }
}

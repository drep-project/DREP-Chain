package network

import (
    "sync"
    "BlockChainTest/common"
    "math/rand"
    "time"
)

var (
    once sync.Once
    singleton *Network
)

type Network struct {
    role int
    miningState int
    channel chan *common.Message
}

func GetInstance(channel chan *common.Message) *Network {
    once.Do(func() {
        singleton = new(Network)
        singleton.channel = channel
    })
    return singleton
}

func (n *Network) Start() int {
    go func() {
        for {
            msg := rand.Intn(3)
            time.Sleep(1 * time.Second)
            switch msg {
            case common.MSG_BLOCK:
                n.channel <- &common.Message{common.MSG_BLOCK, common.Block{rand.Intn(1000), "Block"}}
            case common.MSG_TRANSACTION:
                n.channel <- &common.Message{common.MSG_TRANSACTION, common.Transaction{rand.Intn(1000), "Transaction"}}
            }
        }
    }()
    return 0
}
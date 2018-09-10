package processor

import (
    "BlockChainTest/bean"
    "BlockChainTest/store"
)

type NewComerProcessor struct {
}

func (p *NewComerProcessor) process(msg interface{}) {
    if newcomer, ok := msg.(*bean.Newcomer); ok {
        if user := store.GetItSelfOnUser(); user != nil {
            user.ProcessNewComers(newcomer)
        }
    }
}

type UserProcessor struct {
}

func (p *UserProcessor) process(msg interface{}) {
    if listOfPeer, ok := msg.(*bean.ListOfPeer); ok {
        if newcomer := store.GetItSelfOnNewcomer(); newcomer != nil {
            newcomer.ProcessWelcome(listOfPeer)
        }
    }
}
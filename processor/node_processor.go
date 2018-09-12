package processor

import (
    "BlockChainTest/bean"
    "BlockChainTest/node"
)

type NewComerProcessor struct {
}

func (p *NewComerProcessor) process(msg interface{}) {
    if newcomer, ok := msg.(*bean.Newcomer); ok {
        if bean.IsAdmin {
            node.GetNode().ProcessNewComers(newcomer)
        }
    }
}

type PeersProcessor struct {
}

func (p *PeersProcessor) process(msg interface{}) {
    if listOfPeer, ok := msg.(*bean.ListOfPeer); ok {
        node.GetNode()
    }
}
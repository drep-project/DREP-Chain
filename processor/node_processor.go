package processor

import (
    "BlockChainTest/bean"
    "BlockChainTest/node"
    "BlockChainTest/store"
)

type NewComerProcessor struct {}

func (p *NewComerProcessor) process(msg interface{}) {
    if newcomer, ok := msg.(*bean.Newcomer); ok {
        if store.IsAdmin {
            node.GetNode().ProcessNewComer(newcomer)
        }
    }
}

type PeersProcessor struct {}

func (p *PeersProcessor) process(msg interface{}) {
    if listOfPeer, ok := msg.(*bean.ListOfPeer); ok {
        node.GetNode().ProcessPeers(listOfPeer)
    }
}

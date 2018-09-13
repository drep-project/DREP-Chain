package processor

import (
    "BlockChainTest/bean"
    "BlockChainTest/node"
    "BlockChainTest/store"
)

type NewComerProcessor struct {}

func (p *NewComerProcessor) process(msg interface{}) {
    if peer, ok := msg.(*bean.PeerInfo); ok {
        if store.IsAdmin {
            node.GetNode().ProcessNewPeer(peer)
        }
    }
}

type PeersProcessor struct {}

func (p *PeersProcessor) process(msg interface{}) {
    if list, ok := msg.(*bean.PeerInfoList); ok {
        node.GetNode().ProcessPeerList(list)
    }
}

type MinerProcessor struct {}

func (p *MinerProcessor) process(msg interface{}) {
    // Admin should also process
    if miner, ok := msg.(*bean.MinerInfo); ok {
        node.GetNode().ProcessMinerInfo(miner)
    }
}
package processor

import (
    "BlockChainTest/bean"
    "BlockChainTest/node"
    "fmt"
    "BlockChainTest/network"
    "BlockChainTest/pool"
)

type NewComerProcessor struct {}

func (p *NewComerProcessor) process(peer *network.Peer, msg interface{}) {
    fmt.Println("Receive a new comer 1")
    if peer, ok := msg.(*bean.PeerInfo); ok {
        fmt.Println("Receive a new comer 2", *peer)
        node.GetNode().ProcessNewPeer(peer)
        fmt.Println("Receive a new comer 3")
    }
}

type PeersProcessor struct {}

func (p *PeersProcessor) process(peer *network.Peer, msg interface{}) {
    if list, ok := msg.(*bean.PeerInfoList); ok {
        node.GetNode().ProcessPeerList(list)
    }
}

type BlockReqProcessor struct {}

func (p *BlockReqProcessor) process(peer *network.Peer, msg interface{}) {
    if req, ok := msg.(*bean.BlockReq); ok {
        node.GetNode().ProcessBlockReq(req)
    }
}

type BlockRespProcessor struct {}

func (p *BlockRespProcessor) process(peer *network.Peer, msg interface{}) {
    if resp, ok := msg.(*bean.BlockResp); ok {
        pool.Push(resp)
    }
}

type PingProcessor struct {}

func (p *PingProcessor) process(peer *network.Peer, msg interface{}) {
    if ping, ok := msg.(*bean.Ping); ok {
        node.GetNode().ProcessPing(peer, ping)
    }
}

type PongProcessor struct {}

func (p *PongProcessor) process(peer *network.Peer, msg interface{}) {
    if ping, ok := msg.(*bean.Ping); ok {
        node.GetNode().ProcessPing(peer, ping)
    }
}
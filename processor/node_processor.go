package processor

import (
    "BlockChainTest/bean"
    "BlockChainTest/node"
    "fmt"
)

type NewComerProcessor struct {}

func (p *NewComerProcessor) process(msg interface{}) {
    fmt.Println("Receive a new comer 1")
    if peer, ok := msg.(*bean.PeerInfo); ok {
        fmt.Println("Receive a new comer 2", *peer)
        node.GetNode().ProcessNewPeer(peer)
        fmt.Println("Receive a new comer 3")
    }
}

type PeersProcessor struct {}

func (p *PeersProcessor) process(msg interface{}) {
    if list, ok := msg.(*bean.PeerInfoList); ok {
        node.GetNode().ProcessPeerList(list)
    }
}

type BlockReqProcessor struct {}

func (p *BlockReqProcessor) process(msg interface{}) {
    if req, ok := msg.(*bean.BlockReq); ok {
        node.GetNode().ProcessBlockReq(req)
    }
}

type BlockRespProcessor struct {}

func (p *BlockRespProcessor) process(msg interface{}) {
    if resp, ok := msg.(*bean.BlockResp); ok {
        node.GetNode().ProcessBlockResp(resp)
    }
}
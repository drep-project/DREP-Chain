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

type BlockReqProcessor struct {}

func (p *BlockReqProcessor) process(msg interface{}) {
    if req, ok := msg.(*bean.BlockReq); ok {
        if sender := store.GetItselfOnSender(); sender != nil {
            sender.ProcessBlockReq(req)
        }
    }
}

type BlockRespProcessor struct {}

func (p *BlockRespProcessor) process(msg interface{}) {
    if resp, ok := msg.(*bean.BlockResp); ok {
        if receiver := store.GetItselfOnReceiver(); receiver != nil {
            receiver.ProcessBlockResp(resp)
        }
    }
}
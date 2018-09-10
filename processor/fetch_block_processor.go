package processor

import (
    "BlockChainTest/bean"
    "BlockChainTest/store"
)

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
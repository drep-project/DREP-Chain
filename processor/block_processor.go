package processor

import (
    "fmt"
    "BlockChainTest/common"
    "BlockChainTest/pool"
    "BlockChainTest/util"
    "BlockChainTest/storage"
)

type confirmedBlockProcessor struct {

}

func checkConfirmedBlock(b *common.Block) bool {
    // Check sig
    state := storage.GetState()
    miners := state.GetMiners()
    if !util.Contains(b.Header.LeaderPubKey, miners) {
        return false
    }
    if !util.Subset(b.Header.MinorPubKeys, miners) {
        return false
    }
    return true
}

func (p *confirmedBlockProcessor) process(msg interface{}) {
    if block, ok := msg.(common.Block); ok {
        fmt.Println(block)
        if checkConfirmedBlock(&block) {
            pool.AddBlock(&block)
        }
    }
}

type proposedBlockProcessor struct {

}

func checkProposedBlock(b *common.Block) bool {
    // Check sig
    state := storage.GetState()
    miners := state.GetMiners()
    if !util.Contains(b.Header.LeaderPubKey, miners) {
        return false
    }
    if !util.Subset(b.Header.MinorPubKeys, miners) {
        return false
    }
    return true
}

func (p *proposedBlockProcessor) process(msg interface{}) {
    if block, ok := msg.(common.Block); ok {
        fmt.Println(block)
        if checkProposedBlock(&block) {
            pool.AddBlock(&block)
        }
    }
}

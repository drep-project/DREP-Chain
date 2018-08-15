package processor

import (
    "fmt"
    "BlockChainTest/common"
    "BlockChainTest/pool"
    "BlockChainTest/util"
    "BlockChainTest/storage"
)

type blockProcessor struct {

}

func checkBlock(b *common.Block) bool {
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

func (p *blockProcessor) process(msg interface{}) {
    if block, ok := msg.(common.Block); ok {
        fmt.Println(block)
        if checkBlock(&block) {
            pool.AddBlock(&block)
        }
    }
}
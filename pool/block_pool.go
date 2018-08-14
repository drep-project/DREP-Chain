package pool

import (
    "BlockChainTest/common"
    "sync"
    "log"
)

var (
    blocks map[int32]*common.Block
    blockLock     sync.Mutex
)

func init()  {
    blocks = make(map[int32]*common.Block)
}

func AddBlock(block *common.Block) {
    height := block.Header.Height
    blockLock.Lock()
    if _, exists := blocks[height]; exists {
        log.Fatalf("Block %d exists", height)
    } else {
        blocks[height] = block
    }
    blockLock.Unlock()
}

func GetBlocks(from int32) []*common.Block {
    result := make([]*common.Block, 10)
    blockLock.Lock()
    length := len(blocks)
    count := 0
    for i := from; count <= length; i++ {
        if _, exists := blocks[i]; exists {
            result = append(result, blocks[i])
            delete(blocks, i)
        }
        count++
    }
    blockLock.Unlock()
    return result
}

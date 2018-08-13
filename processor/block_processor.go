package processor

import (
    "fmt"
    "BlockChainTest/common"
)

type blockProcessor struct {

}

func (p *blockProcessor) process(msg interface{})  {
    if block, ok := msg.(common.Block); ok {
        fmt.Println(block)
    }
}
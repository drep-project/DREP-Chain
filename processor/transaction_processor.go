package processor

import (
    "fmt"
    "BlockChainTest/common"
)

type transactionProcessor struct {

}

func (p *transactionProcessor) process(msg interface{})  {
    if transaction, ok := msg.(common.Transaction); ok {
        fmt.Println(transaction)
    }
}
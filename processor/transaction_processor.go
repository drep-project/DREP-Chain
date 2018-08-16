package processor

import (
    "fmt"
    "BlockChainTest/common"
    "BlockChainTest/pool"
)

type transactionProcessor struct {

}

func checkTransaction(t *common.Transaction) bool {
    // Check sig
    return true
}
func (p *transactionProcessor) process(msg interface{})  {
    if transaction, ok := msg.(common.Transaction); ok {
        fmt.Println(transaction)
        if checkTransaction(&transaction) {
            pool.AddTransaction(&transaction)
        }
    }
}
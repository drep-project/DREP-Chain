package processor

import (
    "fmt"
    "BlockChainTest/common"
    "BlockChainTest/pool"
)

type transactionProcessor struct {

}

func checkTransaction(t *common.Transaction) bool {
    // TODO Check sig
    // TODO Check nonce
    return true
}

func transactionExistsInPreviousBlocks(id string) bool {
    return false
}

func (p *transactionProcessor) process(msg interface{})  {
    if transaction, ok := msg.(common.Transaction); ok {
        fmt.Println(transaction)
        id := transaction.GetId()
        if transactionExistsInPreviousBlocks(id) || pool.Contains(id) {
            return
        }
        if checkTransaction(&transaction) {
            pool.AddTransaction(id, &transaction)
            // TODO Send the transaction to all peers
        }
    }
}
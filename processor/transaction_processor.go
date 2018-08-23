package processor

import (
    "fmt"
    "BlockChainTest/pool"
    "BlockChainTest/network"
)

var curve = network.InitCurve()

type transactionProcessor struct {

}

//func checkTransaction(t *common.Transaction) bool {
    // TODO Check sig
    // TODO Check nonce

func checkTransaction(t *network.Transaction) bool {
    // Check sig
    merge, err := t.GetTransactionMerge()
    if err != nil {
        return false
    }
    if !network.Verify(curve, t.Sig, t.PubKey, merge) {
        return false
    }
    return true
}

func transactionExistsInPreviousBlocks(id string) bool {
    return false
}

func (p *transactionProcessor) process(msg interface{})  {
    if transaction, ok := msg.(*network.Transaction); ok {
        fmt.Println(transaction)
        id := transaction.GetId()
        if transactionExistsInPreviousBlocks(id) || pool.Contains(id) {
            return
        }
        //if checkTransaction(&transaction) {
        //    pool.AddTransaction(id, &transaction)
            // TODO Send the transaction to all peers

        if checkTransaction(transaction) {
            pool.AddTransaction(transaction)
        }
    }
}
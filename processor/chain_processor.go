package processor

import (
    "fmt"
    "BlockChainTest/crypto"
    "BlockChainTest/bean"
    "BlockChainTest/store"
    "BlockChainTest/node"
)

var curve = crypto.InitCurve()

type transactionProcessor struct {

}

//func checkTransaction(t *common.Transaction) bool {
    // TODO Check sig
    // TODO Check nonce

func checkTransaction(t *bean.Transaction) bool {
    // Check sig
    //merge, err := t.()
    //if err != nil {
    //    return false
    //}
    //if !network.Verify(curve, t.Sig, t.PubKey, merge) {
    //    return false
    //}
    return true
}

func transactionExistsInPreviousBlocks(id string) bool {
    return false
}

func (p *transactionProcessor) process(msg interface{})  {
    if transaction, ok := msg.(*bean.Transaction); ok {
        fmt.Println(transaction)
        id, _ := transaction.GetId()
        if transactionExistsInPreviousBlocks(id) || store.Contains(id) {
            return
        }
        //if checkTransaction(&transaction) {
        //    pool.AddTransaction(id, &transaction)
            // TODO Send the transaction to all peers

        if checkTransaction(transaction) {
            store.AddTransaction(transaction)
        }
    }
}

type BlockProcessor struct {
    processor *Processor
}

func (p *BlockProcessor) process(msg interface{}) {
    if block, ok := msg.(*bean.Block); ok {
        node.GetNode(nil).ProcessBlock(block)
        p.processor.processRemaining()
    }
}
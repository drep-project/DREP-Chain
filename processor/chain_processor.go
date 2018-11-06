package processor

import (
    "fmt"
    "BlockChainTest/bean"
    "BlockChainTest/store"
    "BlockChainTest/network"
    "BlockChainTest/pool"
)

type transactionProcessor struct {}

//func transactionExistsInPreviousBlocks(id string) bool {
//    return false
//}

func (p *transactionProcessor) process(peer *network.Peer, msg interface{})  {
    if transaction, ok := msg.(*bean.Transaction); ok {
        fmt.Println(transaction)
        id, _ := transaction.TxId()
        if store.ForwardedTransaction(id) {
            fmt.Println("Forwarded this transaction ", *transaction)
            return
        }
        // TODO backup nodes should not add
        if store.AddTransaction(transaction) {
            fmt.Println("Succeed to add this transaction ", *transaction)
            peers := store.GetPeers()
            network.SendMessage(peers, transaction)
            store.ForwardTransaction(id)
        } else {
            fmt.Println("Fail to add this transaction ", *transaction)
        }
    }
}

type BlockProcessor struct {
    processor *Processor
}

func (p *BlockProcessor) process(peer *network.Peer, msg interface{}) {
    if block, ok := msg.(*bean.Block); ok {
        if block.Header.Height <= store.GetCurrentBlockHeight() {
           return
        }
        id, _ := block.BlockID()
        if store.ForwardedBlock(id) { // if forwarded, then processed. later this will be read from db
            fmt.Println("Forwarded this block ", *block)
            return
        }
        store.ForwardBlock(id)
        peers := store.GetPeers()
        network.SendMessage(peers, block)
        pool.Push(block)
        // Here, two blocks will be forwarded here. so is this store enough?
    }
}
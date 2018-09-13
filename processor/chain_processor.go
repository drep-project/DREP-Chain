package processor

import (
    "fmt"
    "BlockChainTest/bean"
    "BlockChainTest/store"
    "BlockChainTest/node"
    "BlockChainTest/network"
)

type transactionProcessor struct {}

//func transactionExistsInPreviousBlocks(id string) bool {
//    return false
//}

func (p *transactionProcessor) process(msg interface{})  {
    if transaction, ok := msg.(*bean.Transaction); ok {
        fmt.Println(transaction)
        id, _ := transaction.TxId()
        if store.Forwarded(id) {
            fmt.Println("Forwarded this transaction ", *transaction)
            return
        }
        if store.GetRole() == bean.OTHER {
            peers := store.GetPeers()
            network.SendMessage(peers, transaction)
            store.Forward(id)
        } else {
            if store.AddTransaction(transaction) {
                fmt.Println("Succeed to add this transaction ", *transaction)
                peers := store.GetPeers()
                network.SendMessage(peers, transaction)
                store.Forward(id)
            } else {
                fmt.Println("Fail to add this transaction ", *transaction)
            }
        }
    }
}

type BlockProcessor struct {
    processor *Processor
}

func (p *BlockProcessor) process(msg interface{}) {
    if block, ok := msg.(*bean.Block); ok {
        if block.Header.Height != store.GetCurrentBlockHeight() + 1 {
            return
        }
        peers := store.GetPeers()
        network.SendMessage(peers, block)
        node.GetNode().ProcessBlock(block, store.GetRole() == bean.MINER)
    }
}
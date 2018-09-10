package processor

import (
    "fmt"
    "BlockChainTest/bean"
    "BlockChainTest/store"
    "BlockChainTest/node"
    "BlockChainTest/network"
)

type transactionProcessor struct {

}

//func transactionExistsInPreviousBlocks(id string) bool {
//    return false
//}

func (p *transactionProcessor) process(msg interface{})  {
    if transaction, ok := msg.(*bean.Transaction); ok {
        fmt.Println(transaction)
        id, _ := transaction.TxId()
        if store.Contains(id) {
            fmt.Println("Contains this transaction ", *transaction)
            return
        }
        if store.AddTransaction(transaction) {
            fmt.Println("Succeed to add this transaction ", *transaction)
            peers := store.GetPeers()
            network.SendMessage(peers, transaction)
        } else {
            fmt.Println("Fail to add this transaction ", *transaction)
        }
    }
}

type BlockProcessor struct {
    processor *Processor
}

func (p *BlockProcessor) process(msg interface{}) {
    if block, ok := msg.(*bean.Block); ok {
        node.GetNode().ProcessBlock(block, true)
    }
}

type NewComerProcessor struct {
}

func (p *NewComerProcessor) process(msg interface{}) {
    if newcomer, ok := msg.(*bean.Newcomer); ok {
        if peer := store.GetItSelfOnPeer(); peer != nil {
            peer.ProcessNewComers(newcomer)
        }
    }
}

type PeerProcessor struct {
}

func (p *PeerProcessor) process(msg interface{}) {
    if listOfNewComer, ok := msg.(*bean.ListOfPeer); ok {
        if newcomer := store.GetItSelfOnPeer(); peer != nil {
            newcomer.ProcessNewComers(listOfNewComer)
        }
    }
}
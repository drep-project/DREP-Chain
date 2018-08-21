package pool

import (
    "BlockChainTest/common"
    "sync"
    "log"
)

var (
    transactions map[string]*common.Transaction
    tranLock     sync.Mutex
)

func init()  {
    transactions = make(map[string]*common.Transaction)
}

func Contains(id string) bool {
    _, exists := transactions[id]
    return exists
}

func AddTransaction(id string, transaction *common.Transaction) {
    tranLock.Lock()
    if _, exists := transactions[id]; exists {
        log.Fatalf("Transaction %s exists", id)
    } else {
        transactions[transaction.GetId()] = transaction
    }
    tranLock.Unlock()
}

func RemoveTransactions(trans []*common.Transaction) {
    tranLock.Lock()
    for _, t := range trans {
        id := t.GetId()
        delete(transactions, id)
    }
    tranLock.Unlock()
}
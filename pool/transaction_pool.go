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
func AddTransaction(transaction *common.Transaction) {
    id := transaction.GetId()
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
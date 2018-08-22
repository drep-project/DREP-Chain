package pool

import (
    "sync"
    "log"
    "BlockChainTest/network"
)

var (
    transactions map[string] *network.Transaction
    tranLock     sync.Mutex
)

func init()  {
    transactions = make(map[string] *network.Transaction)
}


func Contains(id string) bool {
    _, exists := transactions[id]
    return exists
}

//func AddTransaction(id string, transaction *common.Transaction) {
func AddTransaction(transaction *network.Transaction) error {
    id, err := transaction.GetTransactionID()
    if err != nil {
        return err
    }

    tranLock.Lock()
    if _, exists := transactions[id]; exists {
        log.Fatalf("Transaction %s exists", id)
    } else {
        transactions[id] = transaction
    }
    tranLock.Unlock()
    return nil
}

func RemoveTransactions(trans []*network.Transaction) error {
    tranLock.Lock()
    for _, t := range trans {
        id, err := t.GetTransactionID()
        if err != nil {
            return err
        }
        delete(transactions, id)
    }
    tranLock.Unlock()
    return nil
}
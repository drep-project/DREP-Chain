package store

import (
    "sync"
    "log"
    "BlockChainTest/bean"
)

var (
    transactions map[string] *bean.Transaction
    tranLock     sync.Mutex
)

func init()  {
    transactions = make(map[string] *bean.Transaction)
}


func Contains(id string) bool {
    _, exists := transactions[id]
    return exists
}

//func AddTransaction(id string, transaction *common.Transaction) {
func AddTransaction(transaction *bean.Transaction) error {
    id, err := transaction.TxId()
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

func RemoveTransactions(trans []*bean.Transaction) error {
    tranLock.Lock()
    for _, t := range trans {
        id, err := t.TxId()
        if err != nil {
            return err
        }
        delete(transactions, id)
    }
    tranLock.Unlock()
    return nil
}
package store

import (
    "sync"
    "log"
    "BlockChainTest/bean"
    "math/big"
)

type listNode struct {
    tran *bean.Transaction
    prev, next *listNode
}

var (
    tranHead *listNode
    tranTail *listNode
    accountTran map[bean.Address][]*bean.Transaction
    tranSet map[string]bool
    tranLock     sync.Mutex
)

func init()  {
    tranHead = nil
    tranTail = nil
    accountTran = make(map[bean.Address][]*bean.Transaction)
    tranSet = make(map[string]bool)
}


func Contains(id string) bool {
    tranLock.Lock()
    value, exists := tranSet[id]
    if exists && !value {
        delete(tranSet, id)
    }
    tranLock.Unlock()
    return exists || value
}

func checkAndGetAddr(tran *bean.Transaction) (bool, bean.Address) {
    addr := tran.Addr()
    if tran.Data == nil {
        return false, ""
    }
    // TODO Check sig
    if GetNonce(addr) >= tran.Data.Nonce {
        return false, ""
    }
    {
        amount := new(big.Int).SetBytes(tran.Data.Amount)
        gasLimit := new(big.Int).SetBytes(tran.Data.GasLimit)
        gasPrice := new(big.Int).SetBytes(tran.Data.GasPrice)
        total := big.NewInt(0)
        total.Mul(gasLimit, gasPrice)
        total.Add(total, amount)
        if GetBalance(addr).Cmp(total) < 0 {
           return false, ""
        }
    }
    return true, addr
}
//func AddTransaction(id string, transaction *common.Transaction) {
func AddTransaction(transaction *bean.Transaction) {
    check, addr := checkAndGetAddr(transaction)
    if !check {
        return
    }
    id, err := transaction.TxId()
    if err != nil {
        return
    }
    tranLock.Lock()
    if _, exists := tranSet[id]; exists {
        log.Fatalf("Transaction %s exists", id)
    } else {
        tranSet[id] = true
        n := &listNode{tran: transaction}
        if tranHead == nil {
            tranHead = n
            tranTail = n
        } else {
            tranTail.next = n
            n.prev = tranTail
            tranTail = n
        }
        if l, exists := accountTran[addr]; exists {
            l = append(l, transaction)
        } else {
            accountTran[addr] = []*bean.Transaction{transaction}
        }
    }
    tranLock.Unlock()
}

func removeTransaction(tran *bean.Transaction) {
    id, err := tran.TxId()
    if err != nil {
        return
    }
    p := tranHead
    for p != nil {
        tmp, err := p.tran.TxId()
        if err != nil {
            continue
        }
        if id == tmp {
            if p == tranHead && p == tranTail {
                tranHead = nil
                tranTail = nil
            } else if p == tranHead {
                tranHead = p.next
                p.next.prev = nil
                p.next = nil
            } else if p == tranTail {
                tranTail = p.prev
                p.prev.next = nil
                p.prev = nil
            } else {
                p.prev.next = p.next
                p.next.prev = p.prev
                p.next = nil
                p.prev = nil
            }
            return
        }
        p = p.next
    }
    delete(tranSet, id)
    addr := tran.Addr()
    ts := accountTran[addr]
    var i int
    for i = 0; i < len(ts); i++ {
        tmp, err := ts[i].TxId()
        if err != nil {
            continue
        }
        if id == tmp {
            break
        }
    }
    if i < len(ts) {
        accountTran[addr] = append(ts[:i], ts[i + 1:]...)
    }
}

func RemoveTransactions(trans []*bean.Transaction) {
    tranLock.Lock()
    for _, t := range trans {
        removeTransaction(t)
    }
    tranLock.Unlock()
}

func PickTransactions(gasLimit *big.Int) []*bean.Transaction {
    gas := big.NewInt(0)
    tranLock.Lock()
    for p := tranHead; p != nil; p = p.next {

    }
    for _, t := range  {
    }
    tranLock.Unlock()
}
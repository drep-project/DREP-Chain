package store

import (
    "sync"
    "log"
    "BlockChainTest/bean"
    "math/big"
    "BlockChainTest/util/list"
    "bytes"
)

var (
    trans       *list.LinkedList
    accountTran map[bean.Address]*list.SortedLinkedList
    tranSet     map[string]bool
    tranLock    sync.Mutex
    tranCp       = func(a interface{}, b interface{}) int{
        ta, oka := a.(*bean.Transaction)
        tb, okb := b.(*bean.Transaction)
        if oka && okb {
            nonceA := ta.Data.Nonce
            nonceB := tb.Data.Nonce
            if nonceA < nonceB {
                return -1
            } else if nonceA > nonceB {
                return 1
            } else {
                return 0
            }
        } else {
            return 0
        }
    }
)

func init()  {
    trans = list.NewLinkedList()
    accountTran = make(map[bean.Address]*list.SortedLinkedList)
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
func AddTransaction(transaction *bean.Transaction) bool {
    check, addr := checkAndGetAddr(transaction)
    if !check {
        return false
    }
    id, err := transaction.TxId()
    if err != nil {
        return false
    }
    tranLock.Lock()
    if _, exists := tranSet[id]; exists {
        log.Fatalf("Transaction %s exists", id)
        tranLock.Unlock()
        return false
    } else {
        tranSet[id] = true
        trans.Add(transaction)
        if l, exists := accountTran[addr]; exists {
            l.Add(transaction)
        } else {
            l = list.NewSortedLinkedList(tranCp)
            accountTran[addr] = l
            l.Add(transaction)
        }
    }
    tranLock.Unlock()
    return true
}

func removeTransaction(tran *bean.Transaction) {
    id, err := tran.TxId()
    if err != nil {
        return
    }
    cp := func(a interface{}, b interface{}) bool {
        ta, oka := a.(*bean.Transaction)
        tb, okb := b.(*bean.Transaction)
        ba, ea := ta.TxHash()
        bb, eb := tb.TxHash()
        return oka && okb && ea == nil && eb == nil && bytes.Equal(ba, bb)
    }
    trans.Remove(tran, cp)
    delete(tranSet, id)
    addr := tran.Addr()
    ts := accountTran[addr]
    ts.Remove(tran, cp)
}

func RemoveTransactions(trans []*bean.Transaction) {
    tranLock.Lock()
    for _, t := range trans {
        removeTransaction(t)
    }
    tranLock.Unlock()
}

func PickTransactions(gasLimit *big.Int) []*bean.Transaction {
    r := make([]*bean.Transaction, 10)
    gas := big.NewInt(0)
    tranLock.Lock()
    it := trans.Iterator()
    for it.HasNext() {
        if t, ok := it.Next().(*bean.Transaction); ok {
            if id, err := t.TxId(); err != nil {
                if tranSet[id] {
                    addr := t.Addr()
                    if ts, exists := accountTran[addr]; exists {
                        it2 := ts.Iterator()
                        for it2.HasNext() {
                            if t2, ok := it2.Next().(*bean.Transaction); ok {
                                tmp := big.NewInt(0)
                                gasLimit := big.NewInt(0).SetBytes(t2.Data.GasLimit)
                                tmp.Add(gas, gasLimit)
                                if tmp.Cmp(BlockGasLimit) < 0 {
                                    if id2, err := t2.TxId(); err != nil {
                                        gas = tmp
                                        r = append(r, t2)
                                        delete(tranSet, id2)
                                        it2.Remove()
                                    }
                                }
                            }
                        }
                    }
                }
                it.Remove()
            }
        }
    }
    tranLock.Unlock()
    return r
}

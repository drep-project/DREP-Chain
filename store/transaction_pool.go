package store

import (
    "sync"
    "BlockChainTest/log"
    "BlockChainTest/bean"
    "math/big"
    "BlockChainTest/util/list"
    "BlockChainTest/database"
    "BlockChainTest/accounts"
)

var (
    trans       *list.LinkedList
    accountTran map[accounts.CommonAddress]*list.SortedLinkedList
    tranSet     map[string]bool
    tranLock    sync.Mutex
    nonceCp     = func(a interface{}, b interface{}) int{
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
    tranCp = func(a interface{}, b interface{}) bool {
        ta, oka := a.(*bean.Transaction)
        tb, okb := b.(*bean.Transaction)
        sa, ea := ta.TxId()
        sb, eb := tb.TxId()
        return oka && okb && ea == nil && eb == nil && sa == sb
    }
)

func init()  {
    trans = list.NewLinkedList()
    accountTran = make(map[accounts.CommonAddress]*list.SortedLinkedList)
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

func checkAndGetAddr(tran *bean.Transaction) (bool, accounts.CommonAddress) {
    addr := accounts.PubKey2Address(tran.Data.PubKey)
    chainId := tran.Data.ChainId
    if tran.Data == nil {
        return false, accounts.CommonAddress{}
    }
    // TODO Check sig
    if database.GetNonce(addr, chainId) >= tran.Data.Nonce {
        return false, accounts.CommonAddress{}
    }
    {
        amount := new(big.Int).SetBytes(tran.Data.Amount)
        gasLimit := new(big.Int).SetBytes(tran.Data.GasLimit)
        gasPrice := new(big.Int).SetBytes(tran.Data.GasPrice)
        total := big.NewInt(0)
        total.Mul(gasLimit, gasPrice)
        total.Add(total, amount)
        if database.GetBalance(addr, chainId).Cmp(total) < 0 {
            return false, accounts.CommonAddress{}
            // TODO Remove this
        }
    }
    return true, addr
}
//func AddTransaction(id string, transaction *common.transaction) {
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
        log.Error("transaction %s exists", id)
        tranLock.Unlock()
        return false
    } else {
        tranSet[id] = true
        trans.Add(transaction)
        if l, exists := accountTran[addr]; exists {
            l.Add(transaction)
        } else {
            l = list.NewSortedLinkedList(nonceCp)
            accountTran[addr] = l
            l.Add(transaction)
        }
    }
    tranLock.Unlock()
    return true
}

func removeTransaction(tran *bean.Transaction) (bool, bool) {
    id, err := tran.TxId()
    if err != nil {
        return false, false
    }
    r1 := trans.Remove(tran, tranCp)
    delete(tranSet, id)
    addr := accounts.PubKey2Address(tran.Data.PubKey)
    ts := accountTran[addr]
    r2 := ts.Remove(tran, tranCp)
    return r1, r2
}

func PickTransactions(maxGas *big.Int) []*bean.Transaction {
    r := make([]*bean.Transaction, 0) //TODO if 10
    gas := big.NewInt(0)
    tranLock.Lock()
    defer func() {
        tranLock.Unlock()
        for _, t := range r {
            trans.Remove(t, tranCp)
        }
    }()
    it := trans.Iterator()
    tn := make(map[accounts.CommonAddress]int64)
    for it.HasNext() {
        if t, ok := it.Next().(*bean.Transaction); ok {
            if id, err := t.TxId(); err == nil {
                if tranSet[id] {
                    addr := accounts.PubKey2Address(t.Data.PubKey)
                    chainId := t.Data.ChainId
                    if ts, exists := accountTran[addr]; exists {
                        it2 := ts.Iterator()
                        for it2.HasNext() {
                            if t2, ok := it2.Next().(*bean.Transaction); ok {
                                cn, e := tn[addr]
                                if !e {
                                    cn = database.GetNonce(addr, chainId)
                                }
                                if t2.Data.Nonce != cn + 1 {
                                    continue
                                }
                                gasLimit := big.NewInt(0).SetBytes(t2.Data.GasLimit)
                                tmp := big.NewInt(0).Add(gas, gasLimit)
                                if tmp.Cmp(maxGas) <= 0 {
                                    if id2, err := t2.TxId(); err == nil {
                                        gas = tmp
                                        r = append(r, t2)
                                        delete(tranSet, id2)
                                        tn[addr] = t2.Data.Nonce
                                        it2.Remove()
                                    }
                                } else {
                                    return r
                                }
                            } else {
                                it2.Remove()
                            }
                        }
                    } else {
                        log.Error("Fuck")
                    }
                } else {
                    it.Remove()
                }
            }
        } else {
            it.Remove()
        }
    }
    return r
}

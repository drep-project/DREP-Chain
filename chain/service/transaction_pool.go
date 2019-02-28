package service

import (
    "sync"
    chainTypes "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/dlog"
    "github.com/drep-project/drep-chain/crypto"
    "math/big"
    "github.com/drep-project/drep-chain/common/list"
    "github.com/drep-project/drep-chain/database"
    "github.com/pkg/errors"
)

const maxSize = 100000

type TransactionPool struct {
    databaseApi *database.DatabaseService

    trans       *list.LinkedList
    accountTran map[crypto.CommonAddress]*list.SortedLinkedList
    tranSet     map[string]bool
    tranLock    sync.Mutex
    nonceCp     func(a interface{}, b interface{}) int
    tranCp func(a interface{}, b interface{}) bool
}

func NewTransactionPool(databaseApi *database.DatabaseService) *TransactionPool {
 pool := &TransactionPool{databaseApi:databaseApi}
 pool.nonceCp = func(a interface{}, b interface{}) int{
     ta, oka := a.(*chainTypes.Transaction)
     tb, okb := b.(*chainTypes.Transaction)
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
 pool.tranCp = func(a interface{}, b interface{}) bool {
    ta, oka := a.(*chainTypes.Transaction)
    tb, okb := b.(*chainTypes.Transaction)
    sa, ea := ta.TxId()
    sb, eb := tb.TxId()
    return oka && okb && ea == nil && eb == nil && sa == sb
    }
    pool.trans = list.NewLinkedList()
    pool.accountTran = make(map[crypto.CommonAddress]*list.SortedLinkedList)
    pool.tranSet = make(map[string]bool)
 return pool
}

func (pool *TransactionPool) Contains(id string) bool {
    pool.tranLock.Lock()
    defer    pool.tranLock.Unlock()
    value, exists :=  pool.tranSet[id]
    if exists && !value {
        delete( pool.tranSet, id)
    }
    return exists || value
}

func (pool *TransactionPool) checkAndGetAddr(tran *chainTypes.Transaction) (bool, crypto.CommonAddress) {
    addr := crypto.PubKey2Address(tran.Data.PubKey)
    if tran.Data == nil {
        return false, crypto.CommonAddress{}
    }
    // TODO Check sig
    if pool.databaseApi.GetNonce(addr, true) >= tran.Data.Nonce {
        return false, crypto.CommonAddress{}
    }
    {
        amount := new(big.Int).SetBytes(tran.Data.Amount.Bytes())
        gasLimit := new(big.Int).SetBytes(tran.Data.GasLimit.Bytes())
        gasPrice := new(big.Int).SetBytes(tran.Data.GasPrice.Bytes())
        total := big.NewInt(0)
        total.Mul(gasLimit, gasPrice)
        total.Add(total, amount)
        if pool.databaseApi.GetBalance(addr,true).Cmp(total) < 0 {
            return false, crypto.CommonAddress{}
            // TODO Remove this
        }
    }
    return true, addr
}
//func AddTransaction(id string, transaction *common.transaction) {
func (pool *TransactionPool) AddTransaction(transaction *chainTypes.Transaction) error {
    check, addr :=  pool.checkAndGetAddr(transaction)
    if !check {
        return errors.New("check addr failed!")
    }
    id, err := transaction.TxId()
    if err != nil {
        return err
    }
    pool.tranLock.Lock()
    defer  pool.tranLock.Unlock()
    if  pool.trans.Size() >= maxSize {
        msg := "transaction pool full. %s fail to add" + id
        dlog.Error(msg)
        return errors.New(msg)
    }
    if _, exists :=  pool.tranSet[id]; exists {
        msg := "transaction %s exists" + id
        dlog.Error("transaction %s exists", id)
        return errors.New(msg)
    } else {
        pool.tranSet[id] = true
        pool.trans.Add(transaction)
        if l, exists :=  pool.accountTran[addr]; exists {
            l.Add(transaction)
        } else {
            l = list.NewSortedLinkedList( pool.nonceCp)
            pool.accountTran[addr] = l
            l.Add(transaction)
        }
    }
    return nil
}

func (pool *TransactionPool) removeTransaction(tran *chainTypes.Transaction) (bool, bool) {
    id, err := tran.TxId()
    if err != nil {
        return false, false
    }
    pool.tranLock.Lock()
    defer  pool.tranLock.Unlock()
    r1 :=  pool.trans.Remove(tran,  pool.tranCp)
    delete( pool.tranSet, id)
    addr := crypto.PubKey2Address(tran.Data.PubKey)
    ts :=  pool.accountTran[addr]
    r2 := ts.Remove(tran,  pool.tranCp)
    return r1, r2
}

func (pool *TransactionPool) PickTransactions(maxGas *big.Int) []*chainTypes.Transaction {
    r := make([]*chainTypes.Transaction, 0) //TODO if 10
    gas := big.NewInt(0)
    pool.tranLock.Lock()
    defer func() {
        pool.tranLock.Unlock()
        for _, t := range r {
            pool.trans.Remove(t,  pool.tranCp)
        }
    }()
    it :=  pool.trans.Iterator()
    tn := make(map[crypto.CommonAddress]int64)
    for it.HasNext() {
        if t, ok := it.Next().(*chainTypes.Transaction); ok {
            if id, err := t.TxId(); err == nil {
                if  pool.tranSet[id] {
                    addr := crypto.PubKey2Address(t.Data.PubKey)
                    if ts, exists :=  pool.accountTran[addr]; exists {
                        it2 := ts.Iterator()
                        for it2.HasNext() {
                            if t2, ok := it2.Next().(*chainTypes.Transaction); ok {
                                cn, e := tn[addr]
                                if !e {
                                    cn = pool.databaseApi.GetNonce(addr, true)
                                }
                                if t2.Data.Nonce != cn + 1 {
                                    continue
                                }
                                tmp := big.NewInt(0).Add(gas, t2.Data.GasLimit)
                                if tmp.Cmp(maxGas) <= 0 {
                                    if id2, err := t2.TxId(); err == nil {
                                        gas = tmp
                                        r = append(r, t2)
                                        delete( pool.tranSet, id2)
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
                        dlog.Error("Fuck")
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

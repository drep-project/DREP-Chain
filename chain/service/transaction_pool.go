package service

import (
    "sync"
    chainTypes "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/drep-chain/log"
    "github.com/drep-project/drep-chain/crypto"
    "math/big"
    "BlockChainTest/util/list"
    "BlockChainTest/database"
)

const maxSize = 100000

var (
    trans       *list.LinkedList
    accountTran map[crypto.CommonAddress]*list.SortedLinkedList
    tranSet     map[string]bool
    tranLock    sync.Mutex
    nonceCp     = func(a interface{}, b interface{}) int{
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
    tranCp = func(a interface{}, b interface{}) bool {
        ta, oka := a.(*chainTypes.Transaction)
        tb, okb := b.(*chainTypes.Transaction)
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

func checkAndGetAddr(tran *chainTypes.Transaction) (bool, crypto.CommonAddress) {
    addr := crypto.PubKey2Address(tran.Data.PubKey)
    chainId := tran.Data.ChainId
    if tran.Data == nil {
        return false, accounts.CommonAddress{}
    }
    // TODO Check sig
    if database.GetNonce(addr, chainId) >= tran.Data.Nonce {
        return false, accounts.CommonAddress{}
    }
    {
        amount := new(big.Int).SetBytes(tran.Data.Amount.Bytes())
        gasLimit := new(big.Int).SetBytes(tran.Data.GasLimit.Bytes())
        gasPrice := new(big.Int).SetBytes(tran.Data.GasPrice.Bytes())
        total := big.NewInt(0)
        total.Mul(gasLimit, gasPrice)
        total.Add(total, amount)
        if database.GetBalance(addr, chainId).Cmp(total) < 0 {
            return false, crypto.CommonAddress{}
            // TODO Remove this
        }
    }
    return true, addr
}
//func AddTransaction(id string, transaction *common.transaction) {
func AddTransaction(transaction *chainTypes.Transaction) bool {
    check, addr := checkAndGetAddr(transaction)
    if !check {
        return false
    }
    id, err := transaction.TxId()
    if err != nil {
        return false
    }
    tranLock.Lock()
    defer tranLock.Unlock()
    if trans.Size() >= maxSize {
        log.Error("transaction pool full. %s fail to add", id)
        return false
    }
    if _, exists := tranSet[id]; exists {
        log.Error("transaction %s exists", id)
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
    return true
}

func removeTransaction(tran *chainTypes.Transaction) (bool, bool) {
    id, err := tran.TxId()
    if err != nil {
        return false, false
    }
    tranLock.Lock()
    defer tranLock.Unlock()
    r1 := trans.Remove(tran, tranCp)
    delete(tranSet, id)
    addr := crypto.PubKey2Address(tran.Data.PubKey)
    ts := accountTran[addr]
    r2 := ts.Remove(tran, tranCp)
    return r1, r2
}

func PickTransactions(maxGas *big.Int) []*chainTypes.Transaction {
    r := make([]*chainTypes.Transaction, 0) //TODO if 10
    gas := big.NewInt(0)
    tranLock.Lock()
    defer func() {
        tranLock.Unlock()
        for _, t := range r {
            trans.Remove(t, tranCp)
        }
    }()
    it := trans.Iterator()
    tn := make(map[crypto.CommonAddress]int64)
    for it.HasNext() {
        if t, ok := it.Next().(*chainTypes.Transaction); ok {
            if id, err := t.TxId(); err == nil {
                if tranSet[id] {
                    addr := crypto.PubKey2Address(t.Data.PubKey)
                    chainId := t.Data.ChainId
                    if ts, exists := accountTran[addr]; exists {
                        it2 := ts.Iterator()
                        for it2.HasNext() {
                            if t2, ok := it2.Next().(*chainTypes.Transaction); ok {
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

package listenpool

import (
    "sync"
    "BlockChainTest/bean"
    "encoding/json"
    "github.com/drep-project/drep-chain/log"
)
/*
    Ethan
    add ListenPool that BridgeNode use to listen and send Transactions to MainChain
*/


type ListenPool struct {
    lock         sync.Mutex
    cond         *sync.Cond
    Transactions []bean.Transaction
    Size         int
    ChanListen   chan int
}

func NewListenPool() *ListenPool {
    p := &ListenPool{}
    p.cond = sync.NewCond(&p.lock)
    p.Transactions = make([]bean.Transaction, 0)
    p.Size = 0
    p.ChanListen = make(chan int)
    return p
}

func (p *ListenPool) Obtain(cp func(interface{})bool, tranSizeLimit int) []bean.Transaction {
    res := make([]bean.Transaction, 0)
    if (tranSizeLimit >= p.Size) {
        res = p.Transactions[:]
        return res
    }
    tmpSize := p.Size
    i:=len(p.Transactions) - 1
    for tmpSize > tranSizeLimit && i>=0 {
        if bytes,err:=json.Marshal(p.Transactions[i]); err!=nil{
            log.Error("Error, can not json.Marsha(transaction)")
        } else {
            tmpSize-=len(bytes)
            i --
        }
    }
    res = p.Transactions[:i+1]
    return res
}

func (p *ListenPool) Push(tran bean.Transaction)  {
    p.lock.Lock()
    defer p.lock.Unlock()
    p.Transactions = append(p.Transactions, tran)

    if bytes, err := json.Marshal(tran); err!=nil{
        log.Error("Error, can not json.Marsha(transaction)")
    } else {
        p.Size += len(bytes)
    }
    p.cond.Broadcast()
}

func (p *ListenPool) PackageTransaction(transactions []bean.Transaction)  {
    if _,err:=json.Marshal(transactions); err!=nil{
        log.Error("Error, can not json.Marsha(transactions)")
    } else {
        //Todo send bytes to MainChain by using http api
        //when http result come , p.cond.broadcast


    }

}














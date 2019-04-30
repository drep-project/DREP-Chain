package txpool

import (
	"container/heap"
	"errors"
	"fmt"
	"github.com/drep-project/dlog"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"math/big"
	"sync"
)

const maxSize = 100000

//1 池子里的交易按照nonce是否连续，分为乱序的和已经排序的在两个不同的队列中
//2 已经排序好的可以被打包入块
//3 池子里面的交易根据块中的各个地址的交易对应的Nonce进行删除

type TransactionPool struct {
	database *database.Database
	rlock sync.RWMutex
	queue       map[crypto.CommonAddress]*txList
	pending     map[crypto.CommonAddress]*txList
	//accountTran map[crypto.CommonAddress]*list.SortedLinkedList
	allTxs  map[string]bool
	mu      sync.Mutex
	nonceCp func(a interface{}, b interface{}) int
	tranCp  func(a interface{}, b interface{}) bool

	//当前有序的最大的nonce大小,此值应该被存储到DB中（后续考虑txpool的DB存储，一起考虑）
	pendingNonce     map[crypto.CommonAddress]uint64
	eventNewBlockSub event.Subscription
	newBlockChan     chan *chainTypes.Block
	quit             chan struct{}
}

func NewTransactionPool(database *database.Database) *TransactionPool {
	pool := &TransactionPool{database: database}
	pool.nonceCp = func(a interface{}, b interface{}) int {
		ta, oka := a.(*chainTypes.Transaction)
		tb, okb := b.(*chainTypes.Transaction)
		if oka && okb {
			nonceA := ta.Nonce()
			nonceB := tb.Nonce()
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
		sa := ta.TxHash()
		sb := tb.TxHash()
		return oka && okb && sa == sb
	}

	pool.allTxs = make(map[string]bool)
	pool.queue = make(map[crypto.CommonAddress]*txList)
	pool.pending = make(map[crypto.CommonAddress]*txList)
	pool.newBlockChan = make(chan *chainTypes.Block)
	pool.pendingNonce = make(map[crypto.CommonAddress]uint64)

	return pool
}

func (pool *TransactionPool) UpdateState(database *database.Database) {
	pool.rlock.Lock()
	defer pool.rlock.Unlock()
	pool.database = database
}

func (pool *TransactionPool) Contains(id string) bool {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	value, exists := pool.allTxs[id]
	if exists && !value {
		delete(pool.allTxs, id)
	}
	return exists || value
}

//func AddTransaction(id string, transaction *common.transaction) {
func (pool *TransactionPool) AddTransaction(tx *chainTypes.Transaction) error {
	addr, err := tx.From()
	if err != nil {
		return err
	}
	id := tx.TxHash()
	pool.mu.Lock()
	defer pool.mu.Unlock()
	if len(pool.allTxs) >= maxSize {
		msg := fmt.Sprintf("transaction pool full.txid:%s fail to add.pool tx count:%d, maxSize:%d",id, len(pool.allTxs),maxSize)
		dlog.Error(msg)
		return errors.New(msg)
	}

	if _, exists := pool.allTxs[id.String()]; exists {
		msg := "transaction %s exists" + id.String()
		dlog.Error(msg)
		return errors.New(msg)
	} else {
		pool.allTxs[id.String()] = true

		if list, ok := pool.queue[*addr]; ok {
			list.Add(tx)
		} else {
			pool.queue[*addr] = newTxList(true)
			pool.queue[*addr].Add(tx)
		}

		pool.syncToPending(addr)
	}
	return nil
}

func (pool *TransactionPool) syncToPending(address *crypto.CommonAddress) {
	//从queue找nonce连续的交易放入到pending中
	list := pool.queue[*address].Ready(pool.getTransactionCount(address))

	if _, ok := pool.pending[*address]; !ok {
		pool.pending[*address] = newTxList(true)
	}

	var nonce uint64
	listPending := pool.pending[*address]
	if len(list) > 0 {
		for _, tx := range list {
			listPending.Add(tx)
			nonce = tx.Nonce() + 1
		}

		pool.pendingNonce[*address] = nonce
	}
}

func (pool *TransactionPool) removeTransaction(tran *chainTypes.Transaction) (bool, bool) {
	//id, err := tran.TxId()
	//if err != nil {
	//	return false, false
	//}
	//pool.tranLock.Lock()
	//defer pool.tranLock.Unlock()
	//r1 := pool.trans.Remove(tran, pool.tranCp)
	//delete(pool.allTxs, id)
	//addr := crypto.PubKey2Address(tran.Data.PubKey)
	//ts := pool.accountTran[addr]
	//r2 := ts.Remove(tran, pool.tranCp)
	//return r1, r2
	return true, true
}

func (pool *TransactionPool) GetQueue() []*chainTypes.Transaction {
	var retrunTxs []*chainTypes.Transaction

	for _, list := range pool.queue {
		if !list.Empty() {
			txs := list.Flatten()
			retrunTxs = append(retrunTxs, txs...)
		}
	}

	return retrunTxs
}

//打包过程获取交易，进行打包处理
func (pool *TransactionPool) GetPending(GasLimit *big.Int) []*chainTypes.Transaction {
	pool.mu.Lock()
	gasCount := new(big.Int)

	//转数据结构
	hbn := make(map[crypto.CommonAddress]*nonceTxsHeap)
	func() {
		for addr, list := range pool.pending {
			if !list.Empty() {
				txs := list.Flatten()
				newList := &nonceTxsHeap{}
				for _, tx := range txs {
					newList.Push(tx)
				}
				hbn[addr] = newList
			}
		}
	}()
	pool.mu.Unlock()

	var retrunTxs []*chainTypes.Transaction
	for {
		for addr, list := range hbn {
			tx := heap.Pop(list).(*chainTypes.Transaction)
			if GasLimit.Cmp(new(big.Int).Add(tx.GasLimit(), gasCount)) >= 0 {
				retrunTxs = append(retrunTxs, tx)
			} else {
				goto END
			}

			if list.Len() == 0 {
				delete(hbn, addr)
			}
		}

		if len(hbn) <= 0 {
			goto END
		}
	}

END:
	return retrunTxs
}

func (pool *TransactionPool) Start(feed *event.Feed) {
	go pool.checkUpdate()
	pool.eventNewBlockSub = feed.Subscribe(pool.newBlockChan)
}

func (pool *TransactionPool) Stop() {
	close(pool.quit)
	pool.eventNewBlockSub.Unsubscribe()
}

func (pool *TransactionPool) checkUpdate() {
	for {
		select {
		case block := <-pool.newBlockChan:
			pool.adjust(block)
		case <-pool.quit:
			return
		}
	}
}

//已经被处理过NONCE都被清理出去
func (pool *TransactionPool) adjust(block *chainTypes.Block) {
		addrMap := make(map[crypto.CommonAddress]struct{})
		var addrs []*crypto.CommonAddress
		for _, tx := range block.Data.TxList {
			addr, _ := tx.From()
			if _, ok := addrMap[*addr]; !ok {
				addrMap[*addr] = struct{}{}
				addrs = append(addrs, addr)
			}
		}

		if len(addrs) > 0 {
			for addr, _ := range addrMap {
				// 获取数据库里面的nonce
				//根据nonce是否被处理，删除对应的交易
				pool.rlock.RLock()
				nonce := pool.database.GetNonce(&addr)
				pool.rlock.RUnlock()
				pool.mu.Lock()
				list, ok := pool.pending[addr]
				if ok {
					txs := list.Forward(nonce)
					for _, tx := range txs {
						id := tx.TxHash()
						delete(pool.allTxs, id.String())
					}
				}
				pool.mu.Unlock()
				dlog.Warn("clear txpool",  "addr", addr.Hex() ,"max tx.nonce:", nonce,"txpool tx count:", len(pool.allTxs))
			}
		}
}

//获取总的交易个数，即获取地址对应的nonce
func (pool *TransactionPool) GetTransactionCount(address *crypto.CommonAddress) uint64 {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	return pool.getTransactionCount(address)
}

func (pool *TransactionPool) getTransactionCount(address *crypto.CommonAddress) uint64 {
	if nonce, ok := pool.pendingNonce[*address]; ok {
		return nonce
	} else {
		nonce := pool.database.GetNonce(address)
		pool.pendingNonce[*address] = nonce
		return nonce
	}
}

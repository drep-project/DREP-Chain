package txpool

import (
	"container/heap"
	"errors"
	"fmt"
	"github.com/drep-project/drep-chain/chain/store"
	"math/big"
	"sync"
	"time"

	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/types"
)

const (
	maxAllTxsCount  = 100000  //交易池所弄容纳的总的交易数量
	maxTxsOfQueue   = 20      //单个地址对应的乱序队列中，最多容纳交易数目
	maxTxsOfPending = 1000000 //单个地址对应的有序队列中，最多容纳交易数目
	//expireTimeTx    = 60 * 60 * 24 * 7 //交易在一周内，还没有被打包，则被丢弃
	expireTimeTx = 60 * 10 //交易在一周内，还没有被打包，则被丢弃
)

//TransactionPool ...
//1 池子里的交易按照nonce是否连续，分为乱序的和已经排序的在两个不同的队列中
//2 已经排序好的可以被打包入块
//3 池子里面的交易根据块中的各个地址的交易对应的Nonce进行删除
type TransactionPool struct {
	chainStore   store.StoreInterface
	rlock        sync.RWMutex
	queue        map[crypto.CommonAddress]*txList
	pending      map[crypto.CommonAddress]*txList
	allTxs       map[string]*types.Transaction //统计信息使用
	allPricedTxs *txPricedList                 //按照价格排序的tx列表
	mu           sync.Mutex
	nonceCp      func(a interface{}, b interface{}) int
	tranCp       func(a interface{}, b interface{}) bool

	//当前有序的最大的nonce大小,此值应该被存储到DB中（后续考虑txpool的DB存储，一起考虑）
	pendingNonce     map[crypto.CommonAddress]uint64
	eventNewBlockSub event.Subscription
	newBlockChan     chan *types.ChainEvent
	quit             chan struct{}

	// 提供pending交易订阅
	txFeed event.Feed

	//日志
	journal *txJournal
	locals  map[crypto.CommonAddress]struct{} //本地节点包含的地址
}

//NewTransactionPool 创建一个交易池
func NewTransactionPool(chainStore store.StoreInterface, journalPath string) *TransactionPool {
	pool := &TransactionPool{chainStore: chainStore}
	pool.nonceCp = func(a interface{}, b interface{}) int {
		ta, oka := a.(*types.Transaction)
		tb, okb := b.(*types.Transaction)
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
		ta, oka := a.(*types.Transaction)
		tb, okb := b.(*types.Transaction)
		sa := ta.TxHash()
		sb := tb.TxHash()
		return oka && okb && sa == sb
	}

	pool.queue = make(map[crypto.CommonAddress]*txList)
	pool.pending = make(map[crypto.CommonAddress]*txList)
	pool.newBlockChan = make(chan *types.ChainEvent)
	pool.pendingNonce = make(map[crypto.CommonAddress]uint64)

	pool.allTxs = make(map[string]*types.Transaction)
	pool.allPricedTxs = newTxPricedList()

	pool.journal = newTxJournal(journalPath)
	pool.locals = make(map[crypto.CommonAddress]struct{})
	pool.journal.load(pool.addTxs)
	pool.journal.rotate(pool.local())
	//todo 添加本地addr

	return pool
}

func (pool *TransactionPool) journalTx(from crypto.CommonAddress, tx *types.Transaction) {
	// Only journal if it's enabled and the transaction is local
	if _, ok := pool.locals[from]; !ok || pool.journal == nil {
		return
	}

	if err := pool.journal.insert(tx); err != nil {
		log.WithField("Reason", err).Warn("Failed to journal local transaction")
	}
}

func (pool *TransactionPool) local() map[crypto.CommonAddress][]*types.Transaction {
	isLocalAddr := func(addr crypto.CommonAddress) bool {
		_, ok := pool.locals[addr]
		return ok
	}

	all := make(map[crypto.CommonAddress][]*types.Transaction)
	for addr, list := range pool.queue {
		if !list.Empty() && isLocalAddr(addr) {
			txs := list.Flatten()
			all[addr] = txs
		}
	}

	for addr, list := range pool.pending {
		if !list.Empty() && isLocalAddr(addr) {
			txs := list.Flatten()
			if _, ok := all[addr]; ok {
				txs = append(txs, all[addr]...)
			} else {
				all[addr] = txs
			}
		}
	}

	return all
}

func (pool *TransactionPool) addTxs(txs []types.Transaction) []error {
	errs := make([]error, len(txs))
	for i, tx := range txs {
		tx := tx
		from, _ := tx.From()
		if tx.Nonce() < pool.getTransactionCount(from) {
			continue
		}
		errs[i] = pool.addTx(&tx, true)
		if errs[i] != nil {
			log.WithField("Reason", errs[i]).Error("recover tx from journal err")
		}
	}

	return errs
}

//func (pool *TransactionPool) UpdateState(chainStore *chainStore.Database) {
//	pool.rlock.Lock()
//	defer pool.rlock.Unlock()
//	pool.chainStore = chainStore
//}

//func (pool *TransactionPool) Contains(id string) bool {
//	pool.mu.Lock()
//	defer pool.mu.Unlock()
//	_, ok := pool.allTxs[id]
//	//value, exists := pool.allTxs[id]
//	//if exists && !value {
//	//	delete(pool.allTxs, id)
//	//}
//	return ok
//}

//AddTransaction 交易加入到txpool
func (pool *TransactionPool) AddTransaction(tx *types.Transaction, isLocal bool) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return pool.addTx(tx, isLocal)
}

//func AddTransaction(id string, transaction *common.transaction) {
func (pool *TransactionPool) addTx(tx *types.Transaction, isLocal bool) error {
	id := tx.TxHash()
	if _, ok := pool.allTxs[id.String()]; ok {
		return errors.New("konwn tx")
	}

	addr, err := tx.From()
	if err != nil {
		return err
	}

	// pending队列交易替换
	if list, ok := pool.pending[*addr]; ok {
		if list.Overlaps(tx) {
			//替换
			ok, oldTx := list.ReplaceOldTx(tx)
			if !ok {
				return errors.New("can't replace old tx")
			}

			pool.txFeed.Send(types.NewTxsEvent{Txs: []*types.Transaction{tx}})

			log.WithField("nonce", tx.Nonce()).WithField("old price", oldTx.GasPrice()).WithField("new pirce", tx.GasPrice()).Warn("replace")

			delete(pool.allTxs, oldTx.TxHash().String())
			pool.allPricedTxs.Remove(oldTx)

			pool.allTxs[id.String()] = tx
			pool.allPricedTxs.Put(tx)
			pool.journalTx(*addr, tx)
			return nil
		}
	}

	// queue队列交易替换
	if list, ok := pool.queue[*addr]; ok {
		if list.Overlaps(tx) {
			//替换
			ok, oldTx := list.ReplaceOldTx(tx)
			if !ok {
				return errors.New("can't replace old tx")
			}

			log.WithField("nonce", tx.Nonce()).WithField("old price", oldTx.GasPrice()).WithField("new pirce", tx.GasPrice()).Warn("replace")

			delete(pool.allTxs, oldTx.TxHash().String())
			pool.allPricedTxs.Remove(oldTx)

			pool.allTxs[id.String()] = tx
			pool.allPricedTxs.Put(tx)
			pool.journalTx(*addr, tx)
			return nil
		}
	}

	//新的一个交易到来，先看看pool是否满；满的话，删除一些价格较低的tx
	miniPrice := new(big.Int)
	if len(pool.allTxs) >= maxAllTxsCount {
		//todo 价格较低的交易将被丢弃
		txs := pool.allPricedTxs.Discard(1, pool.locals)
		for _, t := range txs {
			if t.GasPrice().Cmp(miniPrice) < 0 || miniPrice.Cmp(new(big.Int)) == 0 {
				miniPrice = t.GasPrice()
			}
			delAddr, _ := tx.From()
			delete(pool.allTxs, t.TxHash().String())

			remove := func(list *txList, pending bool) bool {
				//queue /pending中的tx都要删除掉
				removeSuccess, deleteTxs := list.Remove(&t)
				if removeSuccess {
					for _, delTx := range deleteTxs {
						if pending {
							pool.pendingNonce[*delAddr]--
						}

						delete(pool.allTxs, delTx.TxHash().String())
						pool.allPricedTxs.Remove(delTx)
					}
				}
				return removeSuccess
			}

			for i, maplist := range []map[crypto.CommonAddress]*txList{pool.pending, pool.queue} {
				if list, ok := maplist[*delAddr]; ok {
					b := remove(list, i == 0)
					if b {
						break
					}
				}
			}
		}
	}

	//如果新到来的交易的价格很低，而且不是本地的。那么返回一个错误(需要优化)
	if miniPrice.Cmp(new(big.Int)) != 0 && tx.GasPrice().Cmp(miniPrice) < 0 && !isLocal {
		return fmt.Errorf("new tx gasprice is too low")
	}

	if isLocal {
		if _, ok := pool.locals[*addr]; !ok {
			pool.locals[*addr] = struct{}{}
		}
	}

	//添加到queue
	if list, ok := pool.queue[*addr]; ok {
		//地址对应的队列空间是否已经满 ,删除一些老的tx
		if list.Len() > maxTxsOfQueue {
			//丢弃老的交易
			txs := list.Cap(list.Len())
			for _, delTx := range txs {
				delete(pool.allTxs, delTx.TxHash().String())
				pool.allPricedTxs.Remove(delTx)
			}
		}
		list.Add(tx)
	} else {
		pool.queue[*addr] = newTxList(false)
		pool.queue[*addr].Add(tx)
		pool.txFeed.Send(types.NewTxsEvent{Txs: []*types.Transaction{tx}})
	}

	pool.journalTx(*addr, tx)
	pool.allTxs[id.String()] = tx
	pool.allPricedTxs.Put(tx)
	pool.syncToPending(addr)
	return nil
}

func (pool *TransactionPool) syncToPending(address *crypto.CommonAddress) {
	if _, ok := pool.pending[*address]; !ok {
		pool.pending[*address] = newTxList(true)
	}
	listPending := pool.pending[*address]
	if listPending.Len() > maxTxsOfPending {
		return
	}

	//从queue找nonce连续的交易放入到pending中
	addrList := pool.queue[*address]
	if addrList == nil {
		return
	}
	list := addrList.Ready(pool.getTransactionCount(address))
	var nonce uint64
	if len(list) > 0 {
		for _, tx := range list {
			listPending.Add(tx)
			nonce = tx.Nonce() + 1
		}

		pool.pendingNonce[*address] = nonce
		pool.txFeed.Send(types.NewTxsEvent{Txs: list})
	}
}

//func (pool *TransactionPool) removeTransaction(tran *types.Transaction) (bool, bool) {
//	//id, err := tran.TxId()
//	//if err != nil {
//	//	return false, false
//	//}
//	//pool.tranLock.Lock()
//	//defer pool.tranLock.Unlock()
//	//r1 := pool.trans.Remove(tran, pool.tranCp)
//	//delete(pool.allTxs, id)
//	//addr := crypto.PubKey2Address(tran.Data.PubKey)
//	//ts := pool.accountTran[addr]
//	//r2 := ts.Remove(tran, pool.tranCp)
//	//return r1, r2
//	return true, true
//}

//GetQueue 获取交易池中，非严格排序队列中的所有交易
func (pool *TransactionPool) GetQueue() []*types.Transaction {
	var retrunTxs []*types.Transaction
	pool.mu.Lock()
	defer pool.mu.Unlock()

	for _, list := range pool.queue {
		if !list.Empty() {
			txs := list.Flatten()
			retrunTxs = append(retrunTxs, txs...)
		}
	}

	return retrunTxs
}

//GetPending 打包过程获取交易，进行打包处理
func (pool *TransactionPool) GetPending(GasLimit *big.Int) []*types.Transaction {
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

	var retrunTxs []*types.Transaction
	for {
		for addr, list := range hbn {
			tx := heap.Pop(list).(*types.Transaction)
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

//Start 开启交易池
func (pool *TransactionPool) Start(feed *event.Feed) {
	go pool.checkUpdate()
	pool.eventNewBlockSub = feed.Subscribe(pool.newBlockChan)
}

//Stop 停止交易池
func (pool *TransactionPool) Stop() {
	close(pool.quit)
	pool.eventNewBlockSub.Unsubscribe()
	pool.journal.close()
}

func (pool *TransactionPool) eliminateExpiredTxs() {
	for _, list := range pool.queue {
		if !list.Empty() {
			txs := list.Flatten()
			for _, tx := range txs {
				if tx.Time()+expireTimeTx <= time.Now().Unix() {
					from, _ := tx.From()
					log.WithField("tx time", tx.Time()).WithField("tx nonce", tx.Nonce()).WithField("from", from.String()).Info("tx expire")
					delete(pool.allTxs, tx.TxHash().String())
					pool.allPricedTxs.Remove(tx)
					list.Remove(tx)
				}
			}
		}
	}

	for _, list := range pool.pending {
		if !list.Empty() {
			txs := list.Flatten()
			for _, tx := range txs {
				if tx.Time()+expireTimeTx <= time.Now().Unix() {
					from, _ := tx.From()
					log.WithField("tx time", tx.Time()).WithField("tx nonce", tx.Nonce()).WithField("from", from.String()).Info("tx expire")
					delete(pool.allTxs, tx.TxHash().String())
					pool.allPricedTxs.Remove(tx)
					list.Remove(tx)
				}
			}
		}
	}
}

func (pool *TransactionPool) checkUpdate() {
	timer := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-timer.C:
			pool.mu.Lock()
			//Check whether the transaction is timeout
			pool.eliminateExpiredTxs()

			//to journal
			all := pool.local()
			pool.journal.rotate(all)
			pool.mu.Unlock()
		case block := <-pool.newBlockChan:
			pool.adjust(block.Block)
		case <-pool.quit:
			return
		}
	}
}

//已经被处理过NONCE都被清理出去
func (pool *TransactionPool) adjust(block *types.Block) {
	b := pool.chainStore.RecoverTrie(block.Header.StateRoot)
	if !b {
		log.WithField("recoverRet", b).WithField("h:", block.Header.Height).Error("RecoverTrie")
	}

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
		for addr := range addrMap {
			// 获取数据库里面的nonce
			//根据nonce是否被处理，删除对应的交易
			nonce := pool.chainStore.GetNonce(&addr)
			pool.mu.Lock()
			//块同步的时候，db中的nonce持续增加，pool.pendingNonce[addr]中的值不能被更新
			//此处做更新处理
			if nonce > pool.getTransactionCount(&addr) {
				pool.pendingNonce[addr] = nonce
			}
			list, ok := pool.pending[addr]
			if ok {
				txs := list.Forward(nonce)
				for _, tx := range txs {
					id := tx.TxHash()
					delete(pool.allTxs, id.String())
				}
			}

			pool.syncToPending(&addr)
			pool.mu.Unlock()
			log.WithField("addr", addr.Hex()).WithField("max tx.nonce", nonce).WithField("txpool tx count", len(pool.allTxs)).Warn("clear txpool")
		}
	}
}

//GetTransactionCount 获取总的交易个数，即获取地址对应的nonce
func (pool *TransactionPool) GetTransactionCount(address *crypto.CommonAddress) uint64 {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	return pool.getTransactionCount(address)
}

func (pool *TransactionPool) getTransactionCount(address *crypto.CommonAddress) uint64 {
	if nonce, ok := pool.pendingNonce[*address]; ok {
		return nonce
	}
	nonce := pool.chainStore.GetNonce(address)
	pool.pendingNonce[*address] = nonce
	return nonce
}

//GetTransactions 获取当前池子中所有交易
func (pool *TransactionPool) GetTransactions(addr *crypto.CommonAddress) []types.Transactions {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	twoQueueTxs := make([]types.Transactions, 0, 2)
	if queueList, ok := pool.queue[*addr]; ok {
		queueTxs := make([]types.Transaction, 0, queueList.Len())
		txs := queueList.Flatten()
		for _, tx := range txs {
			queueTxs = append(queueTxs, *tx)
		}
		twoQueueTxs = append(twoQueueTxs, queueTxs)
	}

	if pendingList, ok := pool.pending[*addr]; ok {
		pendingTxs := make([]types.Transaction, 0, pendingList.Len())
		txs := pendingList.Flatten()
		for _, tx := range txs {
			pendingTxs = append(pendingTxs, *tx)
		}
		twoQueueTxs = append(twoQueueTxs, pendingTxs)
	}

	return twoQueueTxs
}

//GetMiniPendingNonce 获取Pending队列中的最小nonce
func (pool *TransactionPool) GetMiniPendingNonce(addr *crypto.CommonAddress) uint64 {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if pendingList, ok := pool.pending[*addr]; ok {
		txs := pendingList.Flatten()
		if len(txs) > 0 {
			return txs[0].Nonce()
		}
	}

	return 0
}

//GetTxInPool 获取交易池中的交易
func (pool *TransactionPool) GetTxInPool(hash string) (*types.Transaction, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if tx, ok := pool.allTxs[hash]; ok {
		return tx, nil
	}
	return nil, fmt.Errorf("hash:%s not in txpool", hash)
}

func (pool *TransactionPool) NewTxFeed() *event.Feed {
	return &pool.txFeed
}

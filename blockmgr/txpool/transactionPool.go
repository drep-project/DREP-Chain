package txpool

import (
	"container/heap"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/drep-project/DREP-Chain/chain/store"

	"github.com/drep-project/DREP-Chain/common/event"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/types"
)

const (
	maxAllTxsCount  = 100000           //The total number of trades held in a trading pool
	maxTxsOfQueue   = 5                //The maximum number of transactions in an out-of-order queue corresponding to a single address
	maxTxsOfPending = 20               //The maximum number of transactions in an ordered queue corresponding to a single address
	expireTimeTx    = 60 * 60 * 24 * 3 //The transaction is discarded if it is not packaged within three days
)

//TransactionPool ...
//1 The transactions in the pool are sorted and sorted in two different queues according to whether or not the nonce is continuous
//2 The sorted ones can be packed into blocks
//3 The transactions in the pool are deleted according to the Nonce corresponding to the transactions of each address in the block
type TransactionPool struct {
	chainStore   store.StoreInterface
	rlock        sync.RWMutex
	queue        map[crypto.CommonAddress]*txList
	pending      map[crypto.CommonAddress]*txList
	allTxs       map[string]*types.Transaction
	allPricedTxs *txPricedList //Tx list sorted by price
	mu           sync.Mutex
	nonceCp      func(a interface{}, b interface{}) int
	tranCp       func(a interface{}, b interface{}) bool

	//The currently ordered maximum nonce size, which should be stored in DB (consider the DB storage of txpool later, together)
	pendingNonce     map[crypto.CommonAddress]uint64
	eventNewBlockSub event.Subscription
	newBlockChan     chan *types.ChainEvent
	quit             chan struct{}

	//Provide pending transaction subscriptions
	txFeed event.Feed

	journal *txJournal
	locals  map[crypto.CommonAddress]struct{} //The address that the local node contains
}

//NewTransactionPool Create a trading pool
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

//AddTransaction ransaction put in txpool
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

	// pending Queue transaction substitution
	if list, ok := pool.pending[*addr]; ok {
		if list.Overlaps(tx) {
			//replace
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

	// Queue transaction substitution
	if list, ok := pool.queue[*addr]; ok {
		if list.Overlaps(tx) {
			//replace
			ok, oldTx := list.ReplaceOldTx(tx)
			if !ok {
				return errors.New("can't replace old tx, new tx price is too low")
			}

			log.WithField("nonce", tx.Nonce()).WithField("old price", oldTx.GasPrice()).WithField("new pirce", tx.GasPrice()).Info("replace")

			delete(pool.allTxs, oldTx.TxHash().String())
			pool.allPricedTxs.Remove(oldTx)

			pool.allTxs[id.String()] = tx
			pool.allPricedTxs.Put(tx)
			pool.journalTx(*addr, tx)
			return nil
		}
	}

	from, err := tx.From()
	nonce := pool.getTransactionCount(from)
	if nonce > tx.Nonce() {
		return fmt.Errorf("SendTransaction local nonce:%d , comming tx nonce:%d too small", nonce, tx.Nonce())
	}

	//A new transaction is coming, let's see if the pool is full; When full, delete some of the cheaper tx's
	miniPrice := new(big.Int)
	if len(pool.allTxs) >= maxAllTxsCount {
		//Cheaper deals will be discarded
		txs := pool.allPricedTxs.Discard(1, pool.locals)
		for _, t := range txs {
			if t.GasPrice().Cmp(miniPrice) < 0 || miniPrice.Cmp(new(big.Int)) == 0 {
				miniPrice = t.GasPrice()
			}
			delAddr, _ := tx.From()
			delete(pool.allTxs, t.TxHash().String())

			remove := func(list *txList, pending bool) bool {
				//going to delete all the tx's in the queue /pending
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

	//If the price of the new transaction is low and not local. So return an error (todo need to optimize)
	if miniPrice.Cmp(new(big.Int)) != 0 && tx.GasPrice().Cmp(miniPrice) < 0 && !isLocal {
		return fmt.Errorf("new tx gasprice is too low")
	}

	if isLocal {
		if _, ok := pool.locals[*addr]; !ok {
			pool.locals[*addr] = struct{}{}
		}
	}

	//add to queue
	if list, ok := pool.queue[*addr]; ok {
		//Whether the queue space corresponding to the address is full,drop old tx
		if list.Len() > maxTxsOfQueue {
			//Blocking here may be nonce discontinuity, so discard the old transaction, add new transaction, and realize nonce continuity
			txs := list.Cap(list.Len())
			for _, delTx := range txs {
				delete(pool.allTxs, delTx.TxHash().String())
				pool.allPricedTxs.Remove(delTx)
				log.WithField("oldtx", delTx.TxHash()).WithField("newTx", tx.TxHash()).Info("old tx been replaced")
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

	//and put it into pending
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

//GetQueue Gets all transactions in the non-strictly sorted queue in the transaction pool
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

//GetPending The packaging process takes the transaction and packages it
func (pool *TransactionPool) GetPending(GasLimit *big.Int) []*types.Transaction {
	pool.mu.Lock()
	gasCount := new(big.Int)

	//change Data structure
	hbn := make(map[crypto.CommonAddress]*nonceTxsHeap)

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

//Start start transaction pool
func (pool *TransactionPool) Start(feed *event.Feed, tipRoot []byte) {
	b := pool.chainStore.RecoverTrie(tipRoot)
	if !b {
		log.WithField("recoverRet", b).Error("tx pool")
	}

	pool.journal.load(pool.addTxs)
	pool.journal.rotate(pool.local())

	go pool.checkUpdate()
	pool.eventNewBlockSub = feed.Subscribe(pool.newBlockChan)
}

//Stop transaction pool work
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

//It's been processed and the NONCE has been cleaned out
func (pool *TransactionPool) adjust(block *types.Block) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

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
			//Get the nonce in the database
			//The corresponding transaction is deleted depending on whether the nonce is processed
			nonce := pool.chainStore.GetNonce(&addr)

			//While the block is synchronized, the nonce in db continues to increase and the value in pool.pendingnonce [addr] cannot be updated
			//Update processing is done here
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
			log.WithField("addr", addr.Hex()).WithField("max tx.nonce", nonce).WithField("txpool tx count", len(pool.allTxs)).Trace("clear txpool")
		}
	}
}

//GetTransactionCount Gets the total number of transactions, that is, the nonce corresponding to the address
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

//GetTransactions Gets all the trades in the current pool
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

//GetMiniPendingNonce Gets the smallest nonce in the Pending queue
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

//GetTxInPool Get transactions in the trading pool
func (pool *TransactionPool) GetTxInPool(hash string) (*types.Transaction, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if tx, ok := pool.allTxs[hash]; ok {
		return tx, nil
	}
	return nil, fmt.Errorf("hash:%s not in txpool", hash)
}

// NewTxFeed new transaction feed in the trading pool
func (pool *TransactionPool) NewTxFeed() *event.Feed {
	return &pool.txFeed
}

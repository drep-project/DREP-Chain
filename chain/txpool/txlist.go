package txpool

import (
	"container/heap"
	"github.com/drep-project/dlog"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"math/big"
	"sort"
)

// nonceHeap is a heap.Interface implementation over 64bit unsigned integers for
// retrieving sorted transactions from the possibly gapped future queue.
type nonceHeap []uint64

func (h nonceHeap) Len() int           { return len(h) }
func (h nonceHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h nonceHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *nonceHeap) Push(x interface{}) {
	*h = append(*h, x.(uint64))
}

func (h *nonceHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// txSortedMap is a nonce->transaction hash map with a heap based index to allow
// iterating over the contents in a nonce-incrementing way.
type txSortedMap struct {
	items map[uint64]*chainTypes.Transaction // Hash map storing the transaction data
	index *nonceHeap                         // Heap of nonces of all the stored transactions (non-strict mode)
	cache []*chainTypes.Transaction          // Cache of the transactions already sorted
}

// newTxSortedMap creates a new nonce-sorted transaction map.
func newTxSortedMap() *txSortedMap {
	return &txSortedMap{
		items: make(map[uint64]*chainTypes.Transaction),
		index: new(nonceHeap),
	}
}

// Get retrieves the current transactions associated with the given nonce.
func (m *txSortedMap) Get(nonce uint64) *chainTypes.Transaction {
	return m.items[nonce]
}

// Put inserts a new transaction into the map, also updating the map's nonce
// index. If a transaction already exists with the same nonce, it's overwritten.
func (m *txSortedMap) Put(tx *chainTypes.Transaction) {
	nonce := tx.Nonce()
	if m.items[nonce] == nil {
		heap.Push(m.index, nonce)
	}
	m.items[nonce], m.cache = tx, nil
}

// Forward removes all transactions from the map with a nonce lower than the
// provided threshold. Every removed transaction is returned for any post-removal
// maintenance.
func (m *txSortedMap) Forward(threshold uint64) []*chainTypes.Transaction {
	var removed []*chainTypes.Transaction

	// Pop off heap items until the threshold is reached
	for m.index.Len() > 0 && (*m.index)[0] < threshold {
		nonce := heap.Pop(m.index).(uint64)
		removed = append(removed, m.items[nonce])
		delete(m.items, nonce)
	}
	// If we had a cached order, shift the front
	if m.cache != nil {
		m.cache = m.cache[len(removed):]
	}
	return removed
}

// Filter iterates over the list of transactions and removes all of them for which
// the specified function evaluates to true.
func (m *txSortedMap) Filter(filter func(*chainTypes.Transaction) bool) []*chainTypes.Transaction {
	var removed []*chainTypes.Transaction

	// Collect all the transactions to filter out
	for nonce, tx := range m.items {
		if filter(tx) {
			removed = append(removed, tx)
			delete(m.items, nonce)
		}
	}
	// If transactions were removed, the heap and cache are ruined
	if len(removed) > 0 {
		*m.index = make([]uint64, 0, len(m.items))
		for nonce := range m.items {
			*m.index = append(*m.index, nonce)
		}
		heap.Init(m.index)

		m.cache = nil
	}
	return removed
}

// Cap places a hard limit on the number of items, returning all transactions
// exceeding that limit.
func (m *txSortedMap) Cap(threshold int) []*chainTypes.Transaction {
	// Short circuit if the number of items is under the limit
	if len(m.items) <= threshold {
		return nil
	}
	// Otherwise gather and drop the highest nonce'd transactions
	var drops []*chainTypes.Transaction

	sort.Sort(*m.index)
	for size := len(m.items); size > threshold; size-- {
		drops = append(drops, m.items[(*m.index)[size-1]])
		delete(m.items, (*m.index)[size-1])
	}
	*m.index = (*m.index)[:threshold]
	heap.Init(m.index)

	// If we had a cache, shift the back
	if m.cache != nil {
		m.cache = m.cache[:len(m.cache)-len(drops)]
	}
	return drops
}

// Remove deletes a transaction from the maintained map, returning whether the
// transaction was found.
func (m *txSortedMap) Remove(nonce uint64) bool {
	// Short circuit if no transaction is present
	_, ok := m.items[nonce]
	if !ok {
		return false
	}
	// Otherwise delete the transaction and fix the heap index
	for i := 0; i < m.index.Len(); i++ {
		if (*m.index)[i] == nonce {
			heap.Remove(m.index, i)
			break
		}
	}
	delete(m.items, nonce)
	m.cache = nil

	return true
}

// Ready retrieves a sequentially increasing list of transactions starting at the
// provided nonce that is ready for processing. The returned transactions will be
// removed from the list.
//
// Note, all transactions with nonces lower than start will also be returned to
// prevent getting into and invalid state. This is not something that should ever
// happen but better to be self correcting than failing!
func (m *txSortedMap) Ready(start uint64) []*chainTypes.Transaction {
	// Short circuit if no transactions are available
	if m.index.Len() == 0 || (*m.index)[0] > start {
		if m.index.Len() != 0 {
			dlog.Warn("txSortedMap Ready", "index[0]:", (*m.index)[0], "req start:", start)
		}
		return nil
	}
	// Otherwise start accumulating incremental transactions
	var ready []*chainTypes.Transaction
	for next := (*m.index)[0]; m.index.Len() > 0 && (*m.index)[0] == next; next++ {
		ready = append(ready, m.items[next])
		delete(m.items, next)
		heap.Pop(m.index)
	}
	m.cache = nil

	return ready
}

// Len returns the length of the transaction map.
func (m *txSortedMap) Len() int {
	return len(m.items)
}

// Flatten creates a nonce-sorted slice of transactions based on the loosely
// sorted internal representation. The result of the sorting is cached in case
// it's requested again before any modifications are made to the contents.
func (m *txSortedMap) Flatten() []*chainTypes.Transaction {
	// If the sorting was not cached yet, create and cache it
	if m.cache == nil {
		m.cache = make([]*chainTypes.Transaction, 0, len(m.items))
		for nonce, tx := range m.items {
			if nonce != tx.Nonce() {
				dlog.Error("call flatten nonce err", "nonce", nonce, "tx.nonoce", tx.Nonce())
				return nil
			}
			m.cache = append(m.cache, tx)
		}
		sort.Sort(nonceTxsHeap(m.cache))
	}
	// Copy the cache to prevent accidental modifications
	txs := make([]*chainTypes.Transaction, len(m.cache))
	copy(txs, m.cache)
	return txs
}

// txList is a "list" of transactions belonging to an account, sorted by account
// nonce. The same type can be used both for storing contiguous transactions for
// the executable/pending queue; and for storing gapped transactions for the non-
// executable/future queue, with minor behavioral changes.
type txList struct {
	strict bool         // Whether nonces are strictly continuous or not
	txs    *txSortedMap // Heap indexed sorted hash map of the transactions
}

// newTxList create a new transaction list for maintaining nonce-indexable fast,
// gapped, sortable transaction lists.
func newTxList(strict bool) *txList {
	return &txList{
		strict: strict,
		txs:    newTxSortedMap(),
	}
}

// Overlaps returns whether the transaction specified has the same nonce as one
// already contained within the list.
func (l *txList) Overlaps(tx *chainTypes.Transaction) bool {
	return l.txs.Get(tx.Nonce()) != nil
}

func (l *txList) ReplaceOldTx(tx *chainTypes.Transaction) (bool, *chainTypes.Transaction) {
	oldTx := l.txs.Get(tx.Nonce())
	if removed := l.txs.Remove(tx.Nonce()); !removed {
		return false, nil
	}
	l.txs.Put(tx)
	return true, oldTx
}

func (l *txList) Add(tx *chainTypes.Transaction) bool {
	old := l.txs.Get(tx.Nonce())
	if old != nil {
		return false
	}

	l.txs.Put(tx)
	return true
}

// Forward removes all transactions from the list with a nonce lower than the
// provided threshold. Every removed transaction is returned for any post-removal
// maintenance.
func (l *txList) Forward(threshold uint64) []*chainTypes.Transaction {
	return l.txs.Forward(threshold)
}

// Filter removes all transactions from the list with a cost or gas limit higher
// than the provided thresholds. Every removed transaction is returned for any
// post-removal maintenance. Strict-mode invalidated transactions are also
// returned.
//
// This method uses the cached costcap and gascap to quickly decide if there's even
// a point in calculating all the costs or if the balance covers all. If the threshold
// is lower than the costgas cap, the caps will be reset to a new high after removing
// the newly invalidated transactions.
func (l *txList) Filter(costLimit *big.Int, gasLimit uint64) ([]chainTypes.Transaction, []chainTypes.Transaction) {
	// If all transactions are below the threshold, short circuit
	//if l.costcap.Cmp(costLimit) <= 0 && l.gascap <= gasLimit {
	//	return nil, nil
	//}
	//l.costcap = new(big.Int).Set(costLimit) // Lower the caps to the thresholds
	//l.gascap = gasLimit
	//
	//// Filter out all the transactions above the account's funds
	//removed := l.txs.Filter(func(tx *types.Transaction) bool { return tx.Cost().Cmp(costLimit) > 0 || tx.Gas() > gasLimit })
	//
	//// If the list was strict, filter anything above the lowest nonce
	//var invalids []types.Transaction
	//
	//if l.strict && len(removed) > 0 {
	//	lowest := uint64(math.MaxUint64)
	//	for _, tx := range removed {
	//		if nonce := tx.Nonce(); lowest > nonce {
	//			lowest = nonce
	//		}
	//	}
	//	invalids = l.txs.Filter(func(tx *types.Transaction) bool { return tx.Nonce() > lowest })
	//}
	//return removed, invalids

	return nil, nil
}

// Cap places a hard limit on the number of items, returning all transactions
// exceeding that limit.
func (l *txList) Cap(threshold int) []*chainTypes.Transaction {
	return l.txs.Cap(threshold)
}

// Remove deletes a transaction from the maintained list, returning whether the
// transaction was found, and also returning any transaction invalidated due to
// the deletion (strict mode only).
func (l *txList) Remove(tx *chainTypes.Transaction) (bool, []*chainTypes.Transaction) {
	// Remove the transaction from the set
	nonce := tx.Nonce()
	if removed := l.txs.Remove(nonce); !removed {
		return false, nil
	}
	// In strict mode, filter out non-executable transactions
	if l.strict {
		return true, l.txs.Filter(func(tx *chainTypes.Transaction) bool { return tx.Nonce() > nonce })
	}
	return true, nil
}

// Ready retrieves a sequentially increasing list of transactions starting at the
// provided nonce that is ready for processing. The returned transactions will be
// removed from the list.
//
// Note, all transactions with nonces lower than start will also be returned to
// prevent getting into and invalid state. This is not something that should ever
// happen but better to be self correcting than failing!
func (l *txList) Ready(start uint64) []*chainTypes.Transaction {
	return l.txs.Ready(start)
}

// Len returns the length of the transaction list.
func (l *txList) Len() int {
	return l.txs.Len()
}

// Empty returns whether the list of transactions is empty or not.
func (l *txList) Empty() bool {
	return l.Len() == 0
}

// Flatten creates a nonce-sorted slice of transactions based on the loosely
// sorted internal representation. The result of the sorting is cached in case
// it's requested again before any modifications are made to the contents.
func (l *txList) Flatten() []*chainTypes.Transaction {
	return l.txs.Flatten()
}

// priceHeap is a heap.Interface implementation over transactions for retrieving
// price-sorted transactions to discard when the pool fills up.
type priceHeap []*chainTypes.Transaction

//
func (h priceHeap) Len() int      { return len(h) }
func (h priceHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h priceHeap) Less(i, j int) bool {
	// Sort primarily by price, returning the cheaper one
	switch h[i].GasPrice().Cmp(h[j].GasPrice()) {
	case -1:
		return true
	case 1:
		return false
	}
	// If the prices match, stabilize via nonces (high nonce is worse)
	return h[i].Nonce() > h[j].Nonce()
}

func (h *priceHeap) Push(x interface{}) {
	*h = append(*h, x.(*chainTypes.Transaction))
}

func (h *priceHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

//txPricedList is a price-sorted heap to allow operating on transactions pool
//contents in a price-incrementing way.
type txPricedList struct {
	items *priceHeap // Heap of prices of all the stored transactions
	//stales int        // Number of stale price points to (re-heap trigger)
}

//newTxPricedList creates a new price-sorted transaction heap.
func newTxPricedList() *txPricedList {
	return &txPricedList{
		items: new(priceHeap),
	}
}

//Put inserts a new transaction into the heap.
func (l *txPricedList) Put(tx *chainTypes.Transaction) {
	heap.Push(l.items, tx)
}

func (l *txPricedList) Remove(tx *chainTypes.Transaction) {
	all := new(priceHeap)
	for len(*l.items) > 0 {
		// Discard stale transactions if found during cleanup
		temp := heap.Pop(l.items).(*chainTypes.Transaction)
		if tx.TxHash() != temp.TxHash() {
			*all = append(*all, temp)
		}
	}

	heap.Init(all)
}

// Discard finds a number of most underpriced transactions, removes them from the
// priced list and returns them for further removal from the entire pool.
func (l *txPricedList) Discard(count int, local map[crypto.CommonAddress]struct{}) chainTypes.Transactions {
	drop := make(chainTypes.Transactions, 0, count) // Remote underpriced transactions to drop
	save := make(chainTypes.Transactions, 0, 64)    // Local underpriced transactions to keep

	for len(*l.items) > 0 && count > 0 {
		// Discard stale transactions if found during cleanup
		tx := heap.Pop(l.items).(*chainTypes.Transaction)
		from, _ := tx.From()
		// Non stale transaction found, discard unless local
		if _, ok := local[*from]; ok {
			save = append(save, *tx)
		} else {
			drop = append(drop, *tx)
			count--
		}
	}
	for _, tx := range save {
		heap.Push(l.items, &tx)
	}
	return drop
}

type TxByNonce []*chainTypes.Transaction

func (s TxByNonce) Len() int           { return len(s) }
func (s TxByNonce) Less(i, j int) bool { return s[i].Nonce() < s[j].Nonce() }
func (s TxByNonce) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type nonceTxsHeap []*chainTypes.Transaction

func (h nonceTxsHeap) Len() int           { return len(h) }
func (h nonceTxsHeap) Less(i, j int) bool { return h[i].Nonce() < h[j].Nonce() }
func (h nonceTxsHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *nonceTxsHeap) Push(x interface{}) {
	*h = append(*h, x.(*chainTypes.Transaction))
}

func (h *nonceTxsHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

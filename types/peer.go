package types

import (
	"container/heap"
	"github.com/drep-project/DREP-Chain/crypto"
	"sync"
	"time"

	//"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/network/p2p"
	"github.com/vishalkuo/bimap"
)

var (
	maxCacheBlockNum = 1024
	maxCacheTxNum    = 1024 //Maximum number of cached transactions per account
)

type PeerInfoInterface interface {
	GetMsgRW() p2p.MsgReadWriter
	GetHeight() uint64
	GetAddr() string

	SetHeight(height uint64)
	KnownTx(tx *Transaction) bool
	MarkTx(tx *Transaction)
	KnownBlock(blk *Block) bool
	MarkBlock(blk *Block)

	SetReqTime(t time.Time)
	CalcAverageRtt()
	AverageRtt() time.Duration
}

var _ PeerInfoInterface = &PeerInfo{}

//Business layer peerknown BLK height:
type PeerInfo struct {
	lock        sync.Mutex
	height      uint64                                //Peer current block height
	exchangeTxs map[crypto.Hash]struct{}              //transaction records exchanged with Peer
	knownTxs    map[crypto.CommonAddress]*sortedBiMap //sorted by NONCE
	knownBlocks *sortedBiMap                          //sorted by height
	peer        *p2p.Peer                             //p2p peer layer
	rw          p2p.MsgReadWriter                     //the protocol corresponding to peer
	reqTime     *time.Time                            //The system time when a request is sent to a peer
	averageRtt  time.Duration                         //The estimated time of the request between local and peer
}

func NewPeerInfo(p *p2p.Peer, rw p2p.MsgReadWriter) *PeerInfo {
	peer := &PeerInfo{
		peer:        p,
		rw:          rw,
		height:      0,
		knownTxs:    make(map[crypto.CommonAddress]*sortedBiMap),
		knownBlocks: newValueSortedBiMap(),
		reqTime:     nil,
		averageRtt:  0,
	}

	return peer
}

func (peer *PeerInfo) SetReqTime(t time.Time) {
	peer.reqTime = &t
}

func (peer *PeerInfo) CalcAverageRtt() {
	duration := time.Since(*peer.reqTime)
	if peer.averageRtt == 0 {
		peer.averageRtt = duration
	} else {
		peer.averageRtt = peer.averageRtt*95/100 + duration*5/100
	}
}

func (peer *PeerInfo) AverageRtt() time.Duration {
	if peer.reqTime != nil && time.Since(*peer.reqTime) > time.Duration(time.Minute*3) {
		peer.averageRtt = 0
	}
	return peer.averageRtt
}

func (peer *PeerInfo) GetAddr() string {
	return peer.peer.IP()
}

//Gets the read-write handle
func (peer *PeerInfo) GetMsgRW() p2p.MsgReadWriter {
	return peer.rw
}

func (peer *PeerInfo) SetHeight(height uint64) {
	peer.height = height
}

func (peer *PeerInfo) GetHeight() uint64 {
	return peer.height
}

//Whether the peer end already knows this tx
func (peer *PeerInfo) KnownTx(tx *Transaction) bool {
	hash := tx.TxHash()
	addr, _ := tx.From()

	peer.lock.Lock()
	defer peer.lock.Unlock()
	if sortedTxs, ok := peer.knownTxs[*addr]; ok {
		if sortedTxs.Exist(hash) {
			return true
		}
	}

	return false
}

//Record the corresponding tx to avoid sending each other multiple times
func (peer *PeerInfo) MarkTx(tx *Transaction) {
	hash := tx.TxHash()
	addr, _ := tx.From()
	peer.lock.Lock()
	defer peer.lock.Unlock()

	if sortedTxs, ok := peer.knownTxs[*addr]; ok {
		if sortedTxs.Len() > maxCacheTxNum {
			sortedTxs.BatchRemove(1)
		}
		sortedTxs.Put(hash, tx.Nonce())
		return
	}

	sortedTxs := newValueSortedBiMap()
	sortedTxs.Put(hash, tx.Nonce())
	peer.knownTxs[*addr] = sortedTxs
}

func (peer *PeerInfo) KnownBlock(blk *Block) bool {
	h := blk.Header.Hash()
	if h == nil {
		return true
	}

	b := peer.knownBlocks.Exist(h)
	if b {
		return b
	}
	return false
}

//Record blocks so that blocks are not synchronized multiple times
func (peer *PeerInfo) MarkBlock(blk *Block) {
	h := blk.Header.Hash()
	if h == nil {
		return
	}

	if peer.knownBlocks.Len() > maxCacheBlockNum {
		peer.knownBlocks.BatchRemove(1)
	}

	peer.knownBlocks.Put(h, blk.Header.Height)

	if peer.height < blk.Header.Height {
		peer.height = blk.Header.Height
	}
}

type uint64SliceHeap []uint64

func (h uint64SliceHeap) Len() int           { return len(h) }
func (h uint64SliceHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h uint64SliceHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *uint64SliceHeap) Push(x interface{}) {
	*h = append(*h, x.(uint64))
}

func (h *uint64SliceHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

//Hash is queued according to the block height corresponding to the hash
type sortedBiMap struct {
	mut   sync.Mutex
	items *bimap.BiMap     //Bidirectional map, key is crypto.hash
	index *uint64SliceHeap //Heap of nonces of all the stored uint64 value
}

// newUint64SortedMap creates a new value-sorted binary map.
func newValueSortedBiMap() *sortedBiMap {
	return &sortedBiMap{
		items: bimap.NewBiMap(),
		index: new(uint64SliceHeap),
	}
}

func (m *sortedBiMap) Exist(key *crypto.Hash) bool {
	return m.items.Exists(*key)
}

func (m *sortedBiMap) GetByKey(key *crypto.Hash) (uint64, bool) {
	m.mut.Lock()
	defer m.mut.Unlock()
	value, ok := m.items.Get(*key)
	if ok {
		return value.(uint64), true
	}
	return 0, false
}

func (m *sortedBiMap) GetByValue(value uint64) *crypto.Hash {
	m.mut.Lock()
	defer m.mut.Unlock()
	hash, ok := m.items.GetInverse(value)
	if ok {
		h := hash.(crypto.Hash)
		return &h
	}
	return nil
}

func (m *sortedBiMap) Put(key *crypto.Hash, value uint64) {
	if _, ok := m.items.Get(*key); !ok {
		m.mut.Lock()
		m.items.Insert(*key, value)
		heap.Push(m.index, value)
		m.mut.Unlock()
	}
}

// Remove deletes a syncHeaderHash from the maintained map, returning whether the
// syncHeaderHash was found.
func (m *sortedBiMap) Remove(value uint64) bool {
	m.mut.Lock()
	defer m.mut.Unlock()
	// Short circuit if no transaction is present
	_, ok := m.items.GetInverse(value)
	if !ok {
		return false
	}
	// Otherwise delete the hash and fix the heap index
	for i := 0; i < m.index.Len(); i++ {
		if (*m.index)[i] == value {
			heap.Remove(m.index, i)
			break
		}
	}
	m.items.DeleteInverse(value)

	return true
}

// Len returns the length of the transaction map.
func (m *sortedBiMap) Len() int {
	return m.items.Size()
}

//According to the value, delete the corresponding k-v from small to large
func (m *sortedBiMap) BatchRemove(count int) int {
	var i int
	for i = 0; i < count && m.items.Size() > 0; i++ {
		m.mut.Lock()
		value := heap.Pop(m.index).(uint64)
		m.items.DeleteInverse(value)

		m.mut.Unlock()
	}

	return i
}

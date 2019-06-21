package types

import (
	"container/heap"
	"github.com/drep-project/drep-chain/crypto"
	"sync"

	//"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/network/p2p"
	"github.com/vishalkuo/bimap"
)

var (
	//DefaultPort      = 55555
	maxCacheBlockNum = 1024
	maxCacheTxNum    = 4096
)

type PeerInfoInterface interface {
	GetMsgRW() p2p.MsgReadWriter
	GetHeight() uint64
}

//业务层peerknown blk height:
type PeerInfo struct {
	height      uint64                   //Peer当前块高度
	exchangeTxs map[crypto.Hash]struct{} //与Peer交换的交易记录
	knownTxs    *sortedBiMap             // 按照NONCE排序
	knownBlocks *sortedBiMap             // 按照高度排序
	peer        *p2p.Peer                //p2p层peer
	rw          p2p.MsgReadWriter        //与peer对应的协议
}

func NewPeerInfo(p *p2p.Peer, rw p2p.MsgReadWriter) *PeerInfo {
	peer := &PeerInfo{
		peer:        p,
		rw:          rw,
		height:      0,
		knownTxs:    newValueSortedBiMap(),
		knownBlocks: newValueSortedBiMap(),
	}

	return peer
}

func (peer *PeerInfo) GetAddr() string {
	return peer.peer.IP()
}

//获取读写句柄
func (peer *PeerInfo) GetMsgRW() p2p.MsgReadWriter {
	return peer.rw
}

func (peer *PeerInfo) SetHeight(height uint64) {
	peer.height = height
}

func (peer *PeerInfo) GetHeight() uint64 {
	return peer.height
}

//peer端是否已经知道此tx
func (peer *PeerInfo) KnownTx(tx *Transaction) bool {
	hash := tx.TxHash()
	b := peer.knownTxs.Exist(hash)
	if b {
		return b
	}
	return false
}

//记录对应的tx，避免多次相互发送
func (peer *PeerInfo) MarkTx(tx *Transaction) {
	hash := tx.TxHash()

	if peer.knownTxs.Len() > maxCacheTxNum {
		peer.knownTxs.BatchRemove(1)
	}

	peer.knownTxs.Put(hash, tx.Nonce())
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

//记录block,以免多次同步块
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

//根据hash对应的块高度，对hash排队
type sortedBiMap struct {
	mut   sync.Mutex
	items *bimap.BiMap     //双向map, key 为 crypto.Hash
	index *uint64SliceHeap // Heap of nonces of all the stored uint64 value
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

//根据value大小，从小到大删除对应的k-v
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

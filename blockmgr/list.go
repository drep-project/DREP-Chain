package blockmgr

import (
	"container/heap"
	"github.com/drep-project/DREP-Chain/crypto"
)

type heightHeap []uint64

func (h heightHeap) Len() int           { return len(h) }
func (h heightHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h heightHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *heightHeap) Push(x interface{}) {
	*h = append(*h, x.(uint64))
}

func (h *heightHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

//根据hash对应的块高度，对hash排队
type heightSortedMap struct {
	items map[uint64]*syncHeaderHash // Hash map storing the syncHeaderHash data
	index *heightHeap                // Heap of nonces of all the stored syncHeaderHash (non-strict mode)
}

// newHeightSortedMap creates a new nonce-sorted syncHeaderHash map.
func newHeightSortedMap() *heightSortedMap {
	return &heightSortedMap{
		items: make(map[uint64]*syncHeaderHash),
		index: new(heightHeap),
	}
}

func (m *heightSortedMap) Get(nonce uint64) *syncHeaderHash {
	return m.items[nonce]
}

func (m *heightSortedMap) Put(shh *syncHeaderHash) {
	height := shh.height
	if _, ok := m.items[height]; !ok {
		m.items[height] = shh
		heap.Push(m.index, height)
	}
}

// Remove deletes a syncHeaderHash from the maintained map, returning whether the
// syncHeaderHash was found.
func (m *heightSortedMap) Remove(height uint64) bool {
	// Short circuit if no transaction is present
	_, ok := m.items[height]
	if !ok {
		return false
	}
	// Otherwise delete the transaction and fix the heap index
	for i := 0; i < m.index.Len(); i++ {
		if (*m.index)[i] == height {
			heap.Remove(m.index, i)
			break
		}
	}
	delete(m.items, height)

	return true
}

// Len returns the length of the transaction map.
func (m *heightSortedMap) Len() int {
	return len(m.items)
}

func (m *heightSortedMap) GetSortedHashs(count int) ([]crypto.Hash, map[crypto.Hash]uint64) {
	hashs := make([]crypto.Hash, 0, count)
	syncHeaderHashs := make(map[crypto.Hash]uint64)

	for i := 0; i < count && m.Len() > 0; i++ {
		h := heap.Pop(m.index).(uint64)
		hashs = append(hashs, *m.items[h].headerHash)
		syncHeaderHashs[*m.items[h].headerHash] = h
		delete(m.items, h)
	}

	return hashs, syncHeaderHashs
}

package types

import (
	"github.com/drep-project/drep-chain/crypto"
	"sync"
)

// BlockIndex provides facilities for keeping track of an in-memory Index of the
// block chain.  Although the name block chain suggests a single chain of
// blocks, it is actually a tree-shaped structure where any node can have
// multiple children.  However, there can only be one active branch which does
// indeed form a chain from the tip all the way back to the genesis block.
type BlockIndex struct {
	sync.RWMutex
	Index map[crypto.Hash]*BlockNode
	Dirty map[*BlockNode]struct{}
}

// newBlockIndex returns a new empty instance of a block Index.  The Index will
// be dynamically populated as block nodes are loaded from the database and
// manually added.
func NewBlockIndex() *BlockIndex {
	return &BlockIndex{
		Index: make(map[crypto.Hash]*BlockNode),
		Dirty: make(map[*BlockNode]struct{}),
	}
}

// HaveBlock returns whether or not the block Index contains the provided hash.
//
// This function is safe for concurrent access.
func (bi *BlockIndex) HaveBlock(hash *crypto.Hash) bool {
	bi.RLock()
	_, hasBlock := bi.Index[*hash]
	bi.RUnlock()
	return hasBlock
}

// LookupNode returns the block node identified by the provided hash.  It will
// return nil if there is no entry for the hash.
//
// This function is safe for concurrent access.
func (bi *BlockIndex) LookupNode(hash *crypto.Hash) *BlockNode {
	bi.RLock()
	node := bi.Index[*hash]
	bi.RUnlock()
	return node
}

// AddNode adds the provided node to the block Index and marks it as Dirty.
// Duplicate entries are not checked so it is up to caller to avoid adding them.
//
// This function is safe for concurrent access.
func (bi *BlockIndex) AddNode(node *BlockNode) {
	bi.Lock()
	bi.addNode(node)
	bi.Dirty[node] = struct{}{}
	bi.Unlock()
}

// addNode adds the provided node to the block Index, but does not mark it as
// Dirty. This can be used while initializing the block Index.
//
// This function is NOT safe for concurrent access.
func (bi *BlockIndex) addNode(node *BlockNode) {
	bi.Index[*node.Hash] = node
}

// NodeStatus provides concurrent-safe access to the status field of a node.
//
// This function is safe for concurrent access.
func (bi *BlockIndex) NodeStatus(node *BlockNode) BlockStatus {
	bi.RLock()
	status := node.Status
	bi.RUnlock()
	return status
}

// SetStatusFlags flips the provided status flags on the block node to on,
// regardless of whether they were on or off previously. This does not unset any
// flags currently on.
//
// This function is safe for concurrent access.
func (bi *BlockIndex) SetStatusFlags(node *BlockNode, flags BlockStatus) {
	bi.Lock()
	node.Status |= flags
	bi.Dirty[node] = struct{}{}
	bi.Unlock()
}

// UnsetStatusFlags flips the provided status flags on the block node to off,
// regardless of whether they were on or off previously.
//
// This function is safe for concurrent access.
func (bi *BlockIndex) UnsetStatusFlags(node *BlockNode, flags BlockStatus) {
	bi.Lock()
	node.Status &^= flags
	bi.Dirty[node] = struct{}{}
	bi.Unlock()
}

// FlushToDB writes all dirty block nodes to the database. If all writes
// succeed, this clears the dirty set.
func (bi *BlockIndex) FlushToDB(storeBlockNodeFunc func (node *BlockNode) error) error {
	bi.Lock()
	if len(bi.Dirty) == 0 {
		bi.Unlock()
		return nil
	}
	var err error
	for node := range bi.Dirty {
		err = storeBlockNodeFunc(node)
		if err != nil {
			break
		}
	}

	// If write was successful, clear the dirty set.
	if err == nil {
		bi.Dirty = make(map[*BlockNode]struct{})
	}

	bi.Unlock()
	return err
}

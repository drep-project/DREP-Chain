package types

import (
	"github.com/drep-project/drep-chain/crypto"
	"math/big"
	"sort"
	"time"
)

var (
	medianTimeBlocks = 11
)

// blockStatus is a bit field representing the validation state of the block.
type BlockStatus byte

const (
	// statusDataStored indicates that the block's payload is stored on disk.
	StatusDataStored BlockStatus = 1 << iota

	// statusValid indicates that the block has been fully validated.
	StatusValid

	// statusValidateFailed indicates that the block has failed validation.
	StatusValidateFailed

	// statusInvalidAncestor indicates that one of the block's ancestors has
	// has failed validation, thus the block is also invalid.
	StatusInvalidAncestor

	// statusNone indicates that the block has no validation state flags set.
	//
	// NOTE: This must be defined last in order to avoid influencing iota.
	StatusNone BlockStatus = 0
)

// HaveData returns whether the full block Data is stored in the database. This
// will return false for a block node where only the header is downloaded or
// kept.
func (status BlockStatus) HaveData() bool {
	return status&StatusDataStored != 0
}

// KnownValid returns whether the block is known to be valid. This will return
// false for a valid block that has not been fully validated yet.
func (status BlockStatus) KnownValid() bool {
	return status&StatusValid != 0
}

// KnownInvalid returns whether the block is known to be invalid. This may be
// because the block itself failed validation or any of its ancestors is
// invalid. This will return false for invalid blocks that have not been proven
// invalid yet.
func (status BlockStatus) KnownInvalid() bool {
	return status&(StatusValidateFailed|StatusInvalidAncestor) != 0
}

type BlockNode struct {
	// NOTE: Additions, deletions, or modifications to the order of the
	// definitions in this struct should not be changed without considering
	// how it affects alignment on 64-bit platforms.  The current order is
	// specifically crafted to result in minimal padding.  There will be
	// hundreds of thousands of these in memory, so a few extra bytes of
	// padding adds up.

	// parent is the parent block for this node.
	Parent *BlockNode `binary:"ignore"`

	// hash is the double sha 256 of the block.
	Hash *crypto.Hash

	StateRoot []byte

	TimeStamp uint64
	// heigh is the position in the block chain.
	Height uint64

	ChainId      ChainIdType
	Version      int32
	PreviousHash *crypto.Hash
	GasLimit     big.Int
	GasUsed      big.Int
	MerkleRoot   []byte
	ReceiptRoot  crypto.Hash
	Bloom        Bloom
	LeaderPubKey crypto.CommonAddress
	MinorPubKeys []crypto.CommonAddress

	Status BlockStatus
}

// initBlockNode initializes a block node from the given header and parent node,
// calculating the height and workSum from the respective fields on the parent.
// This function is NOT safe for concurrent access.  It must only be called when
// initially creating a node.
func InitBlockNode(node *BlockNode, blockHeader *BlockHeader, parent *BlockNode) {
	*node = BlockNode{
		Hash:         blockHeader.Hash(),
		Height:       blockHeader.Height,
		StateRoot:    blockHeader.StateRoot,
		PreviousHash: &blockHeader.PreviousHash,
		TimeStamp:    blockHeader.Timestamp,
		ChainId:      blockHeader.ChainId,
		Version:      blockHeader.Version,
		GasLimit:     blockHeader.GasLimit,
		GasUsed:      blockHeader.GasUsed,
		MerkleRoot:   blockHeader.TxRoot,
		ReceiptRoot:  blockHeader.ReceiptRoot,
		Bloom:        blockHeader.Bloom,
		LeaderPubKey: blockHeader.LeaderAddress,
		MinorPubKeys: blockHeader.MinorAddresses,
	}
	if parent != nil {
		node.Parent = parent
		node.Height = parent.Height + 1
	}
}

// NewBlockNode returns a new block node for the given block header and parent
// node, calculating the height and workSum from the respective fields on the
// parent. This function is NOT safe for concurrent access.
func NewBlockNode(blockHeader *BlockHeader, parent *BlockNode) *BlockNode {
	var node BlockNode
	InitBlockNode(&node, blockHeader, parent)
	return &node
}

// Header constructs a block header from the node and returns it.
//
// This function is safe for concurrent access.
func (node *BlockNode) Header() BlockHeader {
	prevHash := &crypto.Hash{}
	if node.Parent != nil {
		prevHash = node.Parent.Hash
	}
	return BlockHeader{
		Height:         node.Height,
		StateRoot:      node.StateRoot,
		Timestamp:      node.TimeStamp,
		ChainId:        node.ChainId,
		Version:        node.Version,
		PreviousHash:   *prevHash,
		GasLimit:       node.GasLimit,
		GasUsed:        node.GasUsed,
		TxRoot:         node.MerkleRoot,
		ReceiptRoot:    node.ReceiptRoot,
		Bloom:          node.Bloom,
		LeaderAddress:  node.LeaderPubKey,
		MinorAddresses: node.MinorPubKeys,
	}
}

// Ancestor returns the ancestor block node at the provided height by following
// the chain backwards from this node.  The returned block will be nil when a
// height is requested that is after the height of the passed node or is less
// than zero.
//
// This function is safe for concurrent access.
func (node *BlockNode) Ancestor(height uint64) *BlockNode {
	if height < 0 || height > node.Height {
		return nil
	}

	n := node
	for ; n != nil && n.Height != height; n = n.Parent {
		// Intentionally left blank
	}

	return n
}

// RelativeAncestor returns the ancestor block node a relative 'distance' blocks
// before this node.  This is equivalent to calling Ancestor with the node's
// height minus provided distance.
//
// This function is safe for concurrent access.
func (node *BlockNode) RelativeAncestor(distance uint64) *BlockNode {
	return node.Ancestor(node.Height - distance)
}

// CalcPastMedianTime calculates the median time of the previous few blocks
// prior to, and including, the block node.
//
// This function is safe for concurrent access.
func (node *BlockNode) CalcPastMedianTime() time.Time {
	// Create a slice of the previous few block timestamps used to calculate
	// the median per the number defined by the constant medianTimeBlocks.
	timestamps := make([]int64, medianTimeBlocks)
	numNodes := 0
	iterNode := node
	for i := 0; i < medianTimeBlocks && iterNode != nil; i++ {
		timestamps[i] = int64(iterNode.TimeStamp)
		numNodes++

		iterNode = iterNode.Parent
	}

	// Prune the slice to the actual number of available timestamps which
	// will be fewer than desired near the beginning of the block chain
	// and sort them.
	timestamps = timestamps[:numNodes]
	sort.Sort(timeSorter(timestamps))

	// NOTE: The consensus rules incorrectly calculate the median for even
	// numbers of blocks.  A true median averages the middle two elements
	// for a set with an even number of elements in it.   Since the constant
	// for the previous number of blocks to be used is odd, this is only an
	// issue for a few blocks near the beginning of the chain.  I suspect
	// this is an optimization even though the result is slightly wrong for
	// a few of the first blocks since after the first few blocks, there
	// will always be an odd number of blocks in the set per the constant.
	//
	// This code follows suit to ensure the same rules are used, however, be
	// aware that should the medianTimeBlocks constant ever be changed to an
	// even number, this code will be wrong.
	medianTimestamp := timestamps[numNodes/2]
	return time.Unix(medianTimestamp, 0)
}

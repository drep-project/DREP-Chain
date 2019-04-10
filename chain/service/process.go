package service

import (
	"bytes"
	"container/list"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/drep-project/dlog"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"math/big"
	"time"
)

var (
	errBlockExsist       = errors.New("already have block")
	errOrphanBlockExsist = errors.New("already have block (orphan)")
)

func (chainService *ChainService) ProcessGenisisBlock() {
	chainService.ExecuteBlock(chainService.genesisBlock)
	chainService.DatabaseService.RecordBlockJournal(0)
}

func (chainService *ChainService) checkBody(block *chainTypes.Block) error {
	//todo check state root

	// Check whether the block's known, and if not, that it's linkable
	if chainService.blockExists(block.Header.Hash()) {
		return errBlockExsist
	}

	txRoot := chainService.deriveMerkleRoot(block.Data.TxList)
	if !bytes.Equal(txRoot, block.Header.TxRoot) {
		return fmt.Errorf("transaction root hash mismatch: have %x, want %x", string(txRoot), string(block.Header.TxRoot))
	}

	//签名是否正确
	//crypto.ValidateSignatureValues()
	return nil
}

func (cs *ChainService) checkHeader(header *chainTypes.BlockHeader) error {
	//第0个块，不做检查
	if header.Height == 0 {
		return nil
	}
	if cs.Index.HaveBlock(header.Hash()) {
		return nil
	}

	parentBlock, err := cs.DatabaseService.GetBlock(header.PreviousHash)
	if err == nil && header.Height != parentBlock.Header.Height +1 {
		return errors.New("head height err")
	}

	//出块时间
	//if header.Timestamp > time.Now().Unix() {
	//	return fmt.Errorf("block time err", "HeaderTime", time.Unix(header.Timestamp,0), "Now", time.Now())
	//}

	if header.GasLimit.Cmp(new(big.Int).Set(BlockGasLimit)) > 0 {
		return errors.New("gas out block gas limit")
	}

	return nil
}

func (cs *ChainService) checkBlock(block *chainTypes.Block) error {
	// check header
	err := cs.checkHeader(block.Header)
	if err != nil {
		return err
	}
	// check body
	err = cs.checkBody(block)
	if err != nil {
		return err
	}

	//todo check sig
	return nil
}

func (chainService *ChainService) ProcessBlock(block *chainTypes.Block) (bool, bool, error) {
	chainService.addBlockSync.Lock()
	defer chainService.addBlockSync.Unlock()

	err := chainService.checkBlock(block)
	if err != nil {
		dlog.Info("process Block", "err", err)
		return false, false, err
	}

	blockHash := block.Header.Hash()
	exist := chainService.blockExists(blockHash)
	if exist {
		return false, false, errBlockExsist
	}

	// The block must not already exist as an orphan.
	if _, exists := chainService.orphans[*blockHash]; exists {
		return false, false, errOrphanBlockExsist
	}

	// Handle orphan blocks.
	zeroHash := crypto.Hash{}
	prevHash := block.Header.PreviousHash
	prevHashExists := chainService.blockExists(prevHash)
	if !prevHashExists && *prevHash != zeroHash {
		chainService.addOrphanBlock(block)
		return false, true, nil
	}
	isMainChain, err := chainService.acceptBlock(block)
	if err != nil {
		return false, false, err
	}

	// Accept any orphan blocks that depend on this block (they are
	// no longer orphans) and repeat for those accepted blocks until
	// there are no more.
	err = chainService.processOrphans(blockHash)
	if err != nil {
		return false, false, err
	}
	return isMainChain, false, nil
}

func (chainService *ChainService) processOrphans(hash *crypto.Hash) error {
	// Start with processing at least the passed hash.  Leave a little room
	// for additional orphan blocks that need to be processed without
	// needing to grow the array in the common case.
	processHashes := make([]*crypto.Hash, 0, 10)
	processHashes = append(processHashes, hash)
	for len(processHashes) > 0 {
		// Pop the first hash to process from the slice.
		processHash := processHashes[0]
		processHashes[0] = nil // Prevent GC leak.
		processHashes = processHashes[1:]

		for i := 0; i < len(chainService.prevOrphans[*processHash]); i++ {
			orphan := chainService.prevOrphans[*processHash][i]
			if orphan == nil {
				dlog.Warn(fmt.Sprintf("Found a nil entry at index %d in the orphan dependency list for block %v", i, processHash))
				continue
			}

			// Remove the orphan from the orphan pool.
			orphanHash := orphan.Block.Header.Hash()
			chainService.removeOrphanBlock(orphan)
			i--

			// Potentially accept the block into the block chain.
			_, err := chainService.acceptBlock(orphan.Block)
			if err != nil {
				return err
			}

			// Add this block to the list of blocks to process so
			// any orphan blocks that depend on this block are
			// handled too.
			processHashes = append(processHashes, orphanHash)
		}
	}
	return nil
}

func (chainService *ChainService) acceptBlock(block *chainTypes.Block) (bool, error) {
	prevNode := chainService.Index.LookupNode(block.Header.PreviousHash)
	//store block
	err := chainService.DatabaseService.PutBlock(block)
	if err != nil {
		return false, err
	}
	dlog.Info("Accepted block", "Height", block.Header.Height, "Hash", hex.EncodeToString(block.Header.Hash().Bytes()), "TxCount", block.Data.TxCount)
	newNode := chainTypes.NewBlockNode(block.Header, prevNode)
	newNode.Status = chainTypes.StatusDataStored

	chainService.Index.AddNode(newNode)
	err = chainService.Index.FlushToDB(chainService.DatabaseService.PutBlockNode)
	if err != nil {
		return false, err
	}
	if block.Header.PreviousHash.IsEqual(chainService.BestChain.Tip().Hash) {
 		err = chainService.connectBlock(block, newNode)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	if block.Header.Height-chainService.BestChain.Tip().Height <= 0 {
		// store but but not reorg
		dlog.Debug("block store and validate true but not reorgnize")
		return false, nil
	}

	detachNodes, attachNodes := chainService.getReorganizeNodes(newNode)

	// Reorganize the chain.
	dlog.Info("REORGANIZE: Block is causing a reorganize.", "hash", newNode.Hash)
	err = chainService.reorganizeChain(detachNodes, attachNodes)

	// Either getReorganizeNodes or reorganizeChain could have made unsaved
	// changes to the block index, so flush regardless of whether there was an
	// error. The index would only be dirty if the block failed to connect, so
	// we can ignore any errors writing.
	if writeErr := chainService.Index.FlushToDB(chainService.DatabaseService.PutBlockNode); writeErr != nil {
		dlog.Warn("Error flushing block index changes to disk", "Reason", writeErr)
	}
	return err == nil, err
}

func (chainService *ChainService) connectBlock(block *chainTypes.Block, newNode *chainTypes.BlockNode) error {
	//main chain
	if chainService.ValidateBlock(block, chainService.Config.SkipCheckMutiSig||false) {
		chainService.Index.SetStatusFlags(newNode, chainTypes.StatusValid)
	} else {
		chainService.Index.SetStatusFlags(newNode, chainTypes.StatusValidateFailed)
		chainService.flushIndexState()
		return errors.New("validate block fail")
	}
	chainService.flushIndexState()
	_, err := chainService.ExecuteBlock(block)
	if err != nil {
		chainService.Index.SetStatusFlags(newNode, chainTypes.StatusValidateFailed)
		chainService.flushIndexState()
		return err
	}
	chainService.markState(newNode)
	chainService.clearTxPool(block)
	// If this is fast add, or this block node isn't yet marked as
	// valid, then we'll update its status and flush the state to
	// disk again.
	if chainService.Index.NodeStatus(newNode).KnownValid() {
		chainService.Index.SetStatusFlags(newNode, chainTypes.StatusValid)
		chainService.flushIndexState()
	}
	return nil
}

func (chainService *ChainService) flushIndexState() {
	if writeErr := chainService.Index.FlushToDB(chainService.DatabaseService.PutBlockNode); writeErr != nil {
		dlog.Warn("Error flushing block index changes to disk: %v",
			writeErr)
	}
}

func (chainService *ChainService) getReorganizeNodes(node *chainTypes.BlockNode) (*list.List, *list.List) {
	attachNodes := list.New()
	detachNodes := list.New()

	// Do not reorganize to a known invalid chain. Ancestors deeper than the
	// direct parent are checked below but this is a quick check before doing
	// more unnecessary work.
	if chainService.Index.NodeStatus(node.Parent).KnownInvalid() {
		chainService.Index.SetStatusFlags(node, chainTypes.StatusInvalidAncestor)
		return detachNodes, attachNodes
	}

	// Find the fork point (if any) adding each block to the list of nodes
	// to attach to the main tree.  Push them onto the list in reverse order
	// so they are attached in the appropriate order when iterating the list
	// later.
	forkNode := chainService.BestChain.FindFork(node)
	invalidChain := false
	for n := node; n != nil && n != forkNode; n = n.Parent {
		if chainService.Index.NodeStatus(n).KnownInvalid() {
			invalidChain = true
			break
		}
		attachNodes.PushFront(n)
	}

	// If any of the node's ancestors are invalid, unwind attachNodes, marking
	// each one as invalid for future reference.
	if invalidChain {
		var next *list.Element
		for e := attachNodes.Front(); e != nil; e = next {
			next = e.Next()
			n := attachNodes.Remove(e).(*chainTypes.BlockNode)
			chainService.Index.SetStatusFlags(n, chainTypes.StatusInvalidAncestor)
		}
		return detachNodes, attachNodes
	}

	// Start from the end of the main chain and work backwards until the
	// common ancestor adding each block to the list of nodes to detach from
	// the main chain.
	for n := chainService.BestChain.Tip(); n != nil && n != forkNode; n = n.Parent {
		detachNodes.PushBack(n)
	}

	return detachNodes, attachNodes
}

func (chainService *ChainService) reorganizeChain(detachNodes, attachNodes *list.List) error {
	if detachNodes.Len() == 0 && attachNodes.Len() == 0 {
		return nil
	}
	if detachNodes.Len() != 0 {
		elem := detachNodes.Back()
		lastBlock := elem.Value.(*chainTypes.BlockNode)
		height := lastBlock.Height - 1
		chainService.DatabaseService.Rollback2Block(height)
		dlog.Info("REORGANIZE:RollBack state root", "Height", height)
		chainService.markState(lastBlock.Parent)
	}

	if attachNodes.Len() != 0 && detachNodes.Len() != 0 {
		elem := attachNodes.Front()
		for elem != nil { //
			blockNode := elem.Value.(*chainTypes.BlockNode)
			block, err := chainService.DatabaseService.GetBlock(blockNode.Hash)
			if err != nil {
				return err
			}
			err = chainService.connectBlock(block, blockNode)
			if err != nil {
				return err
			}

			dlog.Info("REORGANIZE:Append New Block", "Height", blockNode.Height, "Hash", blockNode.Hash)
			elem = elem.Next()
		}
	}
	return nil
}

func (chainService *ChainService) clearTxPool(block *chainTypes.Block) {
	addrMap := make(map[crypto.CommonAddress]struct{})
	var addrs []*crypto.CommonAddress
	for _, tx := range block.Data.TxList {
		addr := tx.From()
		if _, ok := addrMap[*addr]; !ok {
			addrMap[*addr] = struct{}{}
			addrs = append(addrs, addr)
		}
	}

	if len(addrs) > 0 {
		chainService.newBlockFeed.Send(addrs)
	}
}

func (chainService *ChainService) markState(blockNode *chainTypes.BlockNode) {
	state := chainTypes.NewBestState(blockNode, blockNode.CalcPastMedianTime())
	chainService.DatabaseService.PutChainState(state)
	chainService.BestChain.SetTip(blockNode)
	chainService.stateLock.Lock()
	chainService.StateSnapshot = state
	chainService.stateLock.Unlock()
	chainService.DatabaseService.RecordBlockJournal(state.Height)
}

func (chainService *ChainService) InitStates() error {

	chainState := chainService.DatabaseService.GetChainState()

	var blockCount int32
	err := chainService.DatabaseService.BlockNodeIterator(func(header *chainTypes.BlockHeader, status chainTypes.BlockStatus) error {
		blockCount++
		return nil
	})
	if err != nil {
		return err
	}

	blockNodes := make([]chainTypes.BlockNode, blockCount)

	var i int32
	var lastNode *chainTypes.BlockNode
	err = chainService.DatabaseService.BlockNodeIterator(func(header *chainTypes.BlockHeader, status chainTypes.BlockStatus) error {
		// Determine the parent block node. Since we iterate block headers
		// in order of height, if the blocks are mostly linear there is a
		// very good chance the previous header processed is the parent.
		var parent *chainTypes.BlockNode
		if lastNode == nil {
			blockHash := header.Hash()
			if !blockHash.IsEqual(chainService.genesisBlock.Header.Hash()) {
				return fmt.Errorf("initChainState: Expected  first entry in block index to be genesis block, found %s", blockHash)
			}
		} else if header.PreviousHash == lastNode.Hash {
			// Since we iterate block headers in order of height, if the
			// blocks are mostly linear there is a very good chance the
			// previous header processed is the parent.
			parent = lastNode
		} else {
			parent = chainService.Index.LookupNode(header.PreviousHash)
			if parent == nil {
				return fmt.Errorf(fmt.Sprintf("initChainState: Could not find parent for block %s", header.Hash()))
			}
		}

		// Initialize the block node for the block, connect it,
		// and add it to the block index.
		node := &blockNodes[i]
		chainTypes.InitBlockNode(node, header, parent)
		node.Status = status
		chainService.Index.AddNode(node)

		lastNode = node
		i++
		return nil
	})

	if err != nil {
		return err
	}

	// Set the best chain view to the stored best state.
	tip := chainService.Index.LookupNode(&chainState.Hash)
	if tip == nil {
		return fmt.Errorf(fmt.Sprintf("initChainState: cannot find "+
			"chain tip %s in block index", chainState.Hash))
	}
	chainService.BestChain.SetTip(tip)

	// Load the raw block bytes for the best block.
	_, err = chainService.DatabaseService.GetBlock(&chainState.Hash)
	if err != nil {
		return err
	}

	// As a final consistency check, we'll run through all the
	// nodes which are ancestors of the current chain tip, and mark
	// them as valid if they aren't already marked as such.  This
	// is a safe assumption as all the block before the current tip
	// are valid by definition.
	for iterNode := tip; iterNode != nil; iterNode = iterNode.Parent {
		// If this isn't already marked as valid in the index, then
		// we'll mark it as valid now to ensure consistency once
		// we're up and running.
		if !iterNode.Status.KnownValid() {
			dlog.Info("ancestor of chain tip not marked as valid, upgrading to valid for consistency", "Block", iterNode.Hash, "height", iterNode.Height)
			chainService.Index.SetStatusFlags(iterNode, chainTypes.StatusValid)
		}
	}

	chainService.StateSnapshot = chainTypes.NewBestState(tip, tip.CalcPastMedianTime())

	// As we might have updated the index after it was loaded, we'll
	// attempt to flush the index to the DB. This will only result in a
	// write if the elements are dirty, so it'll usually be a noop.
	return chainService.Index.FlushToDB(chainService.DatabaseService.PutBlockNode)
}

func (chainService *ChainService) createChainState() error {
	node := chainTypes.NewBlockNode(chainService.genesisBlock.Header, nil)
	node.Status = chainTypes.StatusDataStored | chainTypes.StatusValid
	chainService.BestChain.SetTip(node)

	// Add the new node to the index which is used for faster lookups.
	chainService.Index.AddNode(node)

	// Initialize the state related to the best block.  Since it is the
	// genesis block, use its timestamp for the median time.

	chainService.StateSnapshot = chainTypes.NewBestState(node, time.Unix(node.TimeStamp, 0))

	//blockIndexBucketName

	//hashIndexBucketName

	//heightIndexBucketName

	// Save the genesis block to the block index database.
	err := chainService.DatabaseService.PutBlockNode(node)
	if err != nil {
		return err
	}

	// Store the current best chain state into the database.
	err = chainService.DatabaseService.PutChainState(chainService.StateSnapshot)
	if err != nil {
		return err
	}
	err = chainService.DatabaseService.PutBlock(chainService.genesisBlock)
	if err != nil {
		return err
	} else {
		return nil
	}
}

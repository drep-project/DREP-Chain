package chain

import (
	"bytes"
	"container/list"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/database"
	"github.com/drep-project/DREP-Chain/params"
	"github.com/drep-project/DREP-Chain/types"
	"github.com/pkg/errors"
)

func (chainService *ChainService) ProcessBlock(block *types.Block) (bool, bool, error) {
	chainService.addBlockSync.Lock()
	defer chainService.addBlockSync.Unlock()
	blockHash := block.Header.Hash()
	exist := chainService.BlockExists(blockHash)
	if exist {
		return false, false, ErrBlockExsist
	}

	// The block must not already exist as an orphan.
	if _, exists := chainService.orphans[*blockHash]; exists {
		return false, false, ErrOrphanBlockExsist
	}

	// Handle orphan blocks.
	zeroHash := crypto.Hash{}
	prevHash := block.Header.PreviousHash
	prevHashExists := chainService.BlockExists(&prevHash)
	if !prevHashExists && prevHash != zeroHash {
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
				log.Warn(fmt.Sprintf("Found a nil entry at index %d in the orphan dependency list for block %v", i, processHash))
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

func (chainService *ChainService) AcceptBlock(block *types.Block) (inMainChain bool, err error) {
	chainService.addBlockSync.Lock()
	defer chainService.addBlockSync.Unlock()
	return chainService.acceptBlock(block)

}

//TODO cannot find chain tip  对外通知区块失败会产生这个错误  区块未保存 header已经保存
func (chainService *ChainService) acceptBlock(block *types.Block) (inMainChain bool, err error) {
	db := chainService.DatabaseService.BeginTransaction(true)
	prevNode := chainService.blockIndex.LookupNode(&block.Header.PreviousHash)
	preBlock := prevNode.Header()
	for _, blockValidator := range chainService.BlockValidator() {
		err = blockValidator.VerifyHeader(block.Header, &preBlock)
		if err != nil {
			return false, err
		}
		err = blockValidator.VerifyBody(block)
		if err != nil {
			return false, err
		}
	}

	//store block
	err = chainService.DatabaseService.PutBlock(block)
	if err != nil {
		return false, err
	}
	log.WithField("Height", block.Header.Height).WithField("Hash", hex.EncodeToString(block.Header.Hash().Bytes())).WithField("TxCount", block.Data.TxCount).Info("Accepted block")
	newNode := types.NewBlockNode(block.Header, prevNode)
	newNode.Status = types.StatusDataStored

	chainService.blockIndex.AddNode(newNode)
	err = chainService.blockIndex.FlushToDB(chainService.DatabaseService.PutBlockNode)
	if err != nil {
		return false, err
	}

	if block.Header.PreviousHash.IsEqual(chainService.BestChain().Tip().Hash) {
		context, err := chainService.connectBlock(db, block, newNode)
		if err != nil {
			return false, err
		}

		chainService.markState(newNode)
		//SetTip has save tip but block not saving
		chainService.notifyBlock(block, context.Logs)
		return true, nil
	}
	if block.Header.Height <= chainService.BestChain().Tip().Height {
		// store but but not reorg
		log.Debug("block store and validate true but not reorgnize")
		return false, nil
	}

	detachNodes, attachNodes := chainService.getReorganizeNodes(newNode)

	// Reorganize the chain.
	log.WithField("hash", newNode.Hash).Info("REORGANIZE: Block is causing a reorganize.")
	err = chainService.reorganizeChain(db, detachNodes, attachNodes)

	// Either getReorganizeNodes or reorganizeChain could have made unsaved
	// changes to the block index, so flush regardless of whether there was an
	// error. The index would only be dirty if the block failed to connect, so
	// we can ignore any errors writing.
	if writeErr := chainService.blockIndex.FlushToDB(chainService.DatabaseService.PutBlockNode); writeErr != nil {
		log.WithField("Reason", writeErr).Warn("Error flushing block index changes to disk")
	}
	return err == nil, err
}

func (chainService *ChainService) connectBlock(db *database.Database, block *types.Block, newNode *types.BlockNode) (context *BlockExecuteContext, err error) {
	gp := new(GasPool).AddGas(block.Header.GasLimit.Uint64())
	//process transaction
	context = &BlockExecuteContext{
		Db:      db,
		Block:   block,
		Gp:      gp,
		GasUsed: new(big.Int),
		GasFee:  new(big.Int),
	}
	for _, blockValidator := range chainService.BlockValidator() {
		err := blockValidator.ExecuteBlock(context)
		if err != nil {
			break
		}
		//logs = append(logs,allLogs...)
	}

	if err != nil {
		chainService.blockIndex.SetStatusFlags(newNode, types.StatusValidateFailed)
		chainService.flushIndexState()
		return context, err
	}
	if block.Header.GasUsed.Cmp(context.GasUsed) == 0 {
		db.Commit()
		oldStateRoot := db.GetStateRoot()
		if !bytes.Equal(block.Header.StateRoot, oldStateRoot) {
			if !db.RecoverTrie(chainService.bestChain.tip().StateRoot) {
				log.Fatal("root not equal and recover trie err")
			}
			err = errors.Wrapf(ErrNotMathcedStateRoot, "%s not matched %s", hex.EncodeToString(block.Header.StateRoot), hex.EncodeToString(oldStateRoot))
		}
	} else {
		err = errors.Wrapf(ErrGasUsed, "%d not matched %d", block.Header.GasUsed.Uint64(), context.GasUsed.Uint64())
	}

	if err == nil {
		chainService.blockIndex.SetStatusFlags(newNode, types.StatusValid)
		chainService.flushIndexState()
	} else {
		chainService.blockIndex.SetStatusFlags(newNode, types.StatusValidateFailed)
		chainService.flushIndexState()
		return context, err
	}

	// If this is fast add, or this block node isn't yet marked as
	// valid, then we'll update its status and flush the state to
	// disk again.
	if chainService.blockIndex.NodeStatus(newNode).KnownValid() {
		chainService.blockIndex.SetStatusFlags(newNode, types.StatusValid)
		chainService.flushIndexState()
	}
	return context, err
}

func (chainService *ChainService) flushIndexState() {
	if writeErr := chainService.blockIndex.FlushToDB(chainService.DatabaseService.PutBlockNode); writeErr != nil {
		log.WithField("Reason", writeErr).Warn("Error flushing block index changes to disk")
	}
}

func (chainService *ChainService) getReorganizeNodes(node *types.BlockNode) (*list.List, *list.List) {
	attachNodes := list.New()
	detachNodes := list.New()

	// Do not reorganize to a known invalid chain. Ancestors deeper than the
	// direct parent are checked below but this is a quick check before doing
	// more unnecessary work.
	if chainService.blockIndex.NodeStatus(node.Parent).KnownInvalid() {
		chainService.blockIndex.SetStatusFlags(node, types.StatusInvalidAncestor)
		return detachNodes, attachNodes
	}

	// Find the fork point (if any) adding each block to the list of nodes
	// to attach to the main tree.  Push them onto the list in reverse order
	// so they are attached in the appropriate order when iterating the list
	// later.
	forkNode := chainService.BestChain().FindFork(node)
	invalidChain := false
	for n := node; n != nil && n != forkNode; n = n.Parent {
		if chainService.blockIndex.NodeStatus(n).KnownInvalid() {
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
			n := attachNodes.Remove(e).(*types.BlockNode)
			chainService.blockIndex.SetStatusFlags(n, types.StatusInvalidAncestor)
		}
		return detachNodes, attachNodes
	}

	// Start from the end of the main chain and work backwards until the
	// common ancestor adding each block to the list of nodes to detach from
	// the main chain.
	for n := chainService.BestChain().Tip(); n != nil && n != forkNode; n = n.Parent {
		detachNodes.PushBack(n)
	}

	return detachNodes, attachNodes
}

func (chainService *ChainService) reorganizeChain(db *database.Database, detachNodes, attachNodes *list.List) error {
	if detachNodes.Len() == 0 && attachNodes.Len() == 0 {
		return nil
	}
	if detachNodes.Len() != 0 {
		elem := detachNodes.Back()
		lastBlock := elem.Value.(*types.BlockNode)
		height := lastBlock.Height - 1
		db.Rollback2Block(height, lastBlock.Hash)
		log.WithField("Height", height).Info("REORGANIZE:RollBack state root")
		chainService.markState(lastBlock.Parent)
		elem = detachNodes.Front()
		for elem != nil {
			blockNode := elem.Value.(*types.BlockNode)
			block, err := chainService.DatabaseService.GetBlock(blockNode.Hash)
			if err != nil {
				return err
			}
			chainService.notifyDetachBlock(block)
			elem = elem.Next()
		}
	}

	if attachNodes.Len() != 0 && detachNodes.Len() != 0 {
		elem := attachNodes.Front()
		for elem != nil { //
			blockNode := elem.Value.(*types.BlockNode)
			block, err := chainService.DatabaseService.GetBlock(blockNode.Hash)
			if err != nil {
				return err
			}
			context, err := chainService.connectBlock(db, block, blockNode)
			if err != nil {
				return err
			}
			chainService.markState(blockNode)
			chainService.notifyBlock(block, context.Logs)
			log.WithField("Height", blockNode.Height).WithField("Hash", blockNode.Hash).Info("REORGANIZE:Append New Block")
			elem = elem.Next()
		}
	}
	return nil
}

func (chainService *ChainService) notifyBlock(block *types.Block, logs []*types.Log) {
	chainEvent := types.ChainEvent{
		Block: block,
		Hash:  *block.Header.Hash(),
		Logs:  logs,
	}
	chainService.NewBlockFeed().Send(&chainEvent)

	if len(logs) > 0 {
		chainService.logsFeed.Send(logs)
	}
}

func (chainService *ChainService) notifyDetachBlock(block *types.Block) {
	chainService.DetachBlockFeed().Send(block)

	rmLogs := make([]*types.Log, 0)
	receipts := chainService.DatabaseService.GetReceipts(*block.Header.Hash())
	for _, receipt := range receipts {
		for _, log := range receipt.Logs {
			l := *log
			l.Removed = true
			rmLogs = append(rmLogs, &l)
		}
	}
	chainService.rmLogsFeed.Send(types.RemovedLogsEvent{Logs: rmLogs})
}

func (chainService *ChainService) markState(blockNode *types.BlockNode) {
	chainService.BestChain().SetTip(blockNode)
	triedb := chainService.DatabaseService.GetTriedDB()
	triedb.Commit(crypto.Bytes2Hash(blockNode.StateRoot), true)
}

//TODO improves the performan
func (chainService *ChainService) InitStates() error {
	blockCount := chainService.DatabaseService.BlockNodeCount()
	blockNodes := make([]types.BlockNode, blockCount)
	var i int32
	var lastNode *types.BlockNode
	err := chainService.DatabaseService.BlockNodeIterator(func(header *types.BlockHeader, status types.BlockStatus) error {
		// Determine the parent block node. Since we iterate block headers
		// in order of height, if the blocks are mostly linear there is a
		// very good chance the previous header processed is the parent.
		var parent *types.BlockNode
		if lastNode == nil {
			blockHash := header.Hash()
			if !blockHash.IsEqual(chainService.genesisBlock.Header.Hash()) {
				return errors.Wrapf(ErrInitStateFail, "Expected  first entry in block index to be genesis block, found %s", blockHash)
			}
		} else if header.PreviousHash == *lastNode.Hash {
			// Since we iterate block headers in order of height, if the
			// blocks are mostly linear there is a very good chance the
			// previous header processed is the parent.
			parent = lastNode
		} else {
			parent = chainService.blockIndex.LookupNode(&header.PreviousHash)
			if parent == nil {
				return errors.Wrapf(ErrInitStateFail, "Could not find parent for block %s", header.Hash())
			}
		}

		// Initialize the block node for the block, connect it,
		// and add it to the block index.
		node := &blockNodes[i]
		types.InitBlockNode(node, header, parent)
		node.Status = status
		chainService.blockIndex.AddNode(node)

		fmt.Println("get head height:", node.Height)

		lastNode = node
		i++
		return nil
	})

	if err != nil {
		return err
	}

	tip := lastNode
	for {
		if tip.Height != 0 {
			if chainService.DatabaseService.RecoverTrie(tip.StateRoot) {
				break
			}

			// commit fail and repaire here
			//delete dirty data , and rollback state to journalHeight
			//去除磁盘中的nodeblock、block信息；回退区块号及相关的操作日志序列号
			err, _ := chainService.DatabaseService.Rollback2Block(tip.Height, tip.Hash)
			if err != nil {
				log.WithField("height", tip.Height).Error("rollback2block err")
				return err
			}

			//去除内存中节点信息
			chainService.blockIndex.ClearNode(tip)
		} else {
			if chainService.DatabaseService.RecoverTrie(tip.StateRoot) {
				break
			} else {
				return fmt.Errorf("recover tire from old data err")
			}
		}

		tip = tip.Ancestor(tip.Height - 1)
	}

	// Set the best chain view to the stored best state.
	chainService.BestChain().SetTip(tip)

	// Load the raw block bytes for the best block.
	if !chainService.DatabaseService.HasBlock(tip.Hash) {
		return errors.Wrapf(ErrBlockNotFound, "cannot find block %s in block index", tip.Hash)
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
			log.WithField("Block", iterNode.Hash).WithField("height", iterNode.Height).Info("ancestor of chain tip not marked as valid, upgrading to valid for consistency")
			chainService.blockIndex.SetStatusFlags(iterNode, types.StatusValid)
		}
	}

	// As we might have updated the index after it was loaded, we'll
	// attempt to flush the index to the DB. This will only result in a
	// write if the elements are dirty, so it'll usually be a noop.
	return chainService.blockIndex.FlushToDB(chainService.DatabaseService.PutBlockNode)
}

//180000000/360
func (chainService *ChainService) CalcGasLimit(parent *types.BlockHeader, gasFloor, gasCeil uint64) *big.Int {
	limit := uint64(0)
	if parent.GasLimit.Uint64()*2/3 > parent.GasUsed.Uint64() {
		limit = parent.GasLimit.Uint64() - span
	} else {
		limit = parent.GasLimit.Uint64() + span
	}

	if limit < params.MinGasLimit {
		limit = params.MinGasLimit
	}
	// If we're outside our allowed gas range, we try to hone towards them
	if limit < gasFloor {
		limit = gasFloor
	} else if limit > gasCeil {
		limit = gasCeil
	}
	return new(big.Int).SetUint64(limit)
}

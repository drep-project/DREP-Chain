package chainservice

import (
	"bytes"
	"container/list"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/drep-project/drep-chain/chain/params"

	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"github.com/pkg/errors"
)

func (chainService *ChainService) ProcessBlock(block *chainTypes.Block) (bool, bool, error) {
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

func (chainService *ChainService) AcceptBlock(block *chainTypes.Block) (inMainChain bool, err error) {
	chainService.addBlockSync.Lock()
	defer chainService.addBlockSync.Unlock()
	return chainService.acceptBlock(block)

}

//TODO cannot find chain tip  对外通知区块失败会产生这个错误  区块未保存 header已经保存
func (chainService *ChainService) acceptBlock(block *chainTypes.Block) (inMainChain bool, err error) {
	db := chainService.DatabaseService.BeginTransaction()
	defer func() {
		if err == nil {
			db.Commit(true)
		} else {
			db.Discard()
		}
		chainService.blockDb.Commit(false)
	}()
	prevNode := chainService.Index.LookupNode(&block.Header.PreviousHash)
	preBlock := prevNode.Header()
	err = chainService.BlockValidator.VerifyHeader(block.Header, &preBlock)
	if err != nil {
		return false, err
	}
	err = chainService.BlockValidator.VerifyBody(block)
	if err != nil {
		return false, err
	}
	if !chainService.BlockValidator.VerifyMultiSig(block, chainService.Config.SkipCheckMutiSig || false) {
		return false, ErrInvalidateBlockMultisig
	}

	//store block
	err = chainService.blockDb.PutBlock(block)
	if err != nil {
		return false, err
	}
	log.WithField("Height", block.Header.Height).WithField("Hash", hex.EncodeToString(block.Header.Hash().Bytes())).WithField("TxCount", block.Data.TxCount).Info("Accepted block")
	newNode := chainTypes.NewBlockNode(block.Header, prevNode)
	newNode.Status = chainTypes.StatusDataStored

	chainService.Index.AddNode(newNode)
	err = chainService.Index.FlushToDB(chainService.blockDb.PutBlockNode)
	if err != nil {
		return false, err
	}

	if block.Header.PreviousHash.IsEqual(chainService.BestChain.Tip().Hash) {
		err = chainService.connectBlock(db, block, newNode)
		if err != nil {
			return false, err
		}
		chainService.markState(db, newNode)
		//SetTip has save tip but block not saving
		chainService.notifyBlock(block)
		return true, nil
	}
	if block.Header.Height <= chainService.BestChain.Tip().Height {
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
	if writeErr := chainService.Index.FlushToDB(chainService.blockDb.PutBlockNode); writeErr != nil {
		log.WithField("Reason", writeErr).Warn("Error flushing block index changes to disk")
	}
	return err == nil, err
}

func (chainService *ChainService) connectBlock(db *database.Database, block *chainTypes.Block, newNode *chainTypes.BlockNode) (err error) {
	gp := new(GasPool).AddGas(block.Header.GasLimit.Uint64())
	//process transaction
	var gasUsed *big.Int
	var gasFee *big.Int
	gasUsed, gasFee, err = chainService.BlockValidator.ExecuteBlock(db, block, gp)
	if err != nil {
		chainService.Index.SetStatusFlags(newNode, chainTypes.StatusValidateFailed)
		chainService.flushIndexState()
		return err
	}
	err = chainService.AccumulateRewards(db, block, gasFee)
	if err != nil {
		chainService.Index.SetStatusFlags(newNode, chainTypes.StatusValidateFailed)
		chainService.flushIndexState()
		return err
	}
	if block.Header.GasUsed.Cmp(gasUsed) == 0 {
		stateRoot := db.GetStateRoot()
		if !bytes.Equal(block.Header.StateRoot, stateRoot) {
			err = errors.Wrapf(ErrNotMathcedStateRoot, "%s not matched %s", hex.EncodeToString(block.Header.StateRoot), hex.EncodeToString(stateRoot))
		}
	} else {
		err = errors.Wrapf(ErrGasUsed, "%d not matched %d", block.Header.GasUsed.Uint64(), gasUsed.Uint64())
	}

	if err == nil {
		chainService.Index.SetStatusFlags(newNode, chainTypes.StatusValid)
		chainService.flushIndexState()
	} else {
		chainService.Index.SetStatusFlags(newNode, chainTypes.StatusValidateFailed)
		chainService.flushIndexState()
		return err
	}

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
	if writeErr := chainService.Index.FlushToDB(chainService.blockDb.PutBlockNode); writeErr != nil {
		log.WithField("Reason",writeErr).Warn("Error flushing block index changes to disk")
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

func (chainService *ChainService) reorganizeChain(db *database.Database, detachNodes, attachNodes *list.List) error {
	if detachNodes.Len() == 0 && attachNodes.Len() == 0 {
		return nil
	}
	if detachNodes.Len() != 0 {
		elem := detachNodes.Back()
		lastBlock := elem.Value.(*chainTypes.BlockNode)
		height := lastBlock.Height - 1
		db.Rollback2Block(height)
		log.WithField("Height", height).Info("REORGANIZE:RollBack state root")
		chainService.markState(db, lastBlock.Parent)
		elem = detachNodes.Front()
		for elem != nil {
			blockNode := elem.Value.(*chainTypes.BlockNode)
			block, err := chainService.blockDb.GetBlock(blockNode.Hash)
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
			blockNode := elem.Value.(*chainTypes.BlockNode)
			block, err := chainService.blockDb.GetBlock(blockNode.Hash)
			if err != nil {
				return err
			}
			err = chainService.connectBlock(db, block, blockNode)
			if err != nil {
				return err
			}
			chainService.markState(db, blockNode)
			chainService.notifyBlock(block)
			log.WithField( "Height", blockNode.Height).WithField("Hash", blockNode.Hash).Info("REORGANIZE:Append New Block")
			elem = elem.Next()
		}
	}
	return nil
}

func (chainService *ChainService) notifyBlock(block *chainTypes.Block) {
	chainService.NewBlockFeed.Send(block)
}

func (chainService *ChainService) notifyDetachBlock(block *chainTypes.Block) {
	chainService.DetachBlockFeed.Send(block)
}

func (chainService *ChainService) markState(db *database.Database, blockNode *chainTypes.BlockNode) {
	state := chainTypes.NewBestState(blockNode)
	chainService.BestChain.SetTip(blockNode)
	chainService.stateLock.Lock()
	chainService.StateSnapshot = &ChainState{
		BestState: *state,
		db:        db,
	}
	chainService.stateLock.Unlock()
	chainService.DatabaseService.PutChainState(state)
	db.Commit(true)
	db.RecordBlockJournal(state.Height)
	chainService.blockDb.Commit(false)
}

//TODO improves the performan
func (chainService *ChainService) InitStates() error {
	chainState := chainService.DatabaseService.GetChainState()
	journalHeight := chainService.DatabaseService.GetBlockJournal()
	if chainState.Height > journalHeight {
		if chainState.Height-journalHeight == 1 {
			// commit fail and repaire here
			//delete dirty data , and rollback state to journalHeight
			chainService.DatabaseService.Rollback2Block(journalHeight)
			header, _, err := chainService.DatabaseService.GetBlockNode(&chainState.PrevHash, journalHeight)
			if err != nil {
				return err
			}
			node := chainTypes.NewBlockNode(header, nil)
			chainState = chainTypes.NewBestState(node)
			chainService.DatabaseService.PutChainState(chainState)
			log.Info("Repair commit fail")
		} else {
			panic("never reach here")
		}
	}
	_, err := chainService.blockDb.GetBlock(&chainState.Hash)
	if err != nil {
		//block not save but tip save status is ok
		rollbackHeight :=  chainState.Height - 1
		chainService.DatabaseService.Rollback2Block(rollbackHeight)
		header, _, err := chainService.DatabaseService.GetBlockNode(&chainState.PrevHash, rollbackHeight)
		if err != nil {
			return err
		}
		node := chainTypes.NewBlockNode(header, nil)
		chainState = chainTypes.NewBestState(node)
		chainService.DatabaseService.PutChainState(chainState)
		log.Info("Repair block missing")
	}

	blockCount := chainService.DatabaseService.BlockNodeCount()
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
				return errors.Wrapf(ErrInitStateFail, "Expected  first entry in block index to be genesis block, found %s", blockHash)
			}
		} else if header.PreviousHash == *lastNode.Hash {
			// Since we iterate block headers in order of height, if the
			// blocks are mostly linear there is a very good chance the
			// previous header processed is the parent.
			parent = lastNode
		} else {
			parent = chainService.Index.LookupNode(&header.PreviousHash)
			if parent == nil {
				return errors.Wrapf(ErrInitStateFail, "Could not find parent for block %s", header.Hash())
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
		return errors.Wrapf(ErrInitStateFail, "cannot find chain tip %s in block index", chainState.Hash)
	}
	chainService.BestChain.SetTip(tip)

	// Load the raw block bytes for the best block.
	if !chainService.DatabaseService.HasBlock(&chainState.Hash) {
		return errors.Wrapf(ErrBlockNotFound, "cannot find block %s in block index", chainState.Hash)
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
			chainService.Index.SetStatusFlags(iterNode, chainTypes.StatusValid)
		}
	}
	chainService.stateLock.Lock()
	chainService.StateSnapshot = &ChainState{
		BestState: *chainTypes.NewBestState(tip),
		db:        chainService.DatabaseService.BeginTransaction(),
	}
	chainService.stateLock.Unlock()

	// As we might have updated the index after it was loaded, we'll
	// attempt to flush the index to the DB. This will only result in a
	// write if the elements are dirty, so it'll usually be a noop.
	return chainService.Index.FlushToDB(chainService.DatabaseService.PutBlockNode)
}

//180000000/360
func (chainService *ChainService) CalcGasLimit(parent *chainTypes.BlockHeader, gasFloor, gasCeil uint64) *big.Int {
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

// AccumulateRewards credits,The leader gets half of the reward and other ,Other participants get the average of the other half
func (chainService *ChainService) AccumulateRewards(db *database.Database, b *chainTypes.Block, totalGasBalance *big.Int) error {
	reward := new(big.Int).SetUint64(uint64(params.Rewards))
	leaderAddr := crypto.PubKey2Address(&b.Header.LeaderPubKey)

	r := new(big.Int)
	r = r.Div(reward, new(big.Int).SetInt64(2))
	r.Add(r, totalGasBalance)
	err := db.AddBalance(&leaderAddr, r)
	if err != nil {
		return err
	}
	num := len(b.Header.MinorPubKeys)
	for _, memberPK := range b.Header.MinorPubKeys {
		if !memberPK.IsEqual(&b.Header.LeaderPubKey) {
			memberAddr := crypto.PubKey2Address(&memberPK)
			r.Div(reward, new(big.Int).SetInt64(int64(num*2)))
			err = db.AddBalance(&memberAddr, r)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

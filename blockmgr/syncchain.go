package blockmgr

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/types"

	"github.com/drep-project/DREP-Chain/common/event"
	"github.com/drep-project/DREP-Chain/crypto"
)

type tasksTxsSync struct {
	txs  []*types.Transaction
	peer *types.PeerInfo
}

//同步缓存池中的交易
func (blockMgr *BlockMgr) syncTxs() {
	for {
		select {
		case task := <-blockMgr.taskTxsCh:
			count := len(task.txs) / maxTxsCount
			var i int
			for i = 0; i < count; i++ {
				blockMgr.P2pServer.Send(task.peer.GetMsgRW(), types.MsgTypeTransaction, task.txs[i*maxTxsCount:(i+1)*maxTxsCount])
				time.Sleep(time.Millisecond * maxSyncSleepTime)
			}

			blockMgr.P2pServer.Send(task.peer.GetMsgRW(), types.MsgTypeTransaction, task.txs[count*maxTxsCount:])
		case <-blockMgr.quit:
			return
		}
	}
}

func (blockMgr *BlockMgr) synchronise() {
	timer := time.NewTicker(time.Second * 10)
	defer timer.Stop()

	syncBlock := func() {
		pi := blockMgr.GetBestPeerInfo()
		if pi == nil {
			return
		}
		currentHeight := blockMgr.ChainService.BestChain().Height()
		if pi.GetHeight() > currentHeight {
			log.Info("need sync  ", pi.GetHeight(), ">", currentHeight)
			err := blockMgr.fetchBlocks(pi)
			if err != nil {
				log.WithField("Reason", err).Warn("sync block from peer")
			}
		}
	}

	syncTx := func(peer *types.PeerInfo) {
		//保证能把pending里面的所有tx全部取出来
		txs := blockMgr.transactionPool.GetPending(new(big.Int).SetUint64(0xffffffffffffffff))
		txs2 := blockMgr.transactionPool.GetQueue()

		txs = append(txs, txs2...)
		blockMgr.taskTxsCh <- tasksTxsSync{peer: peer, txs: append(txs, txs2...)}
	}

	for {
		select {
		case <-timer.C:
			syncBlock()
		case peer := <-blockMgr.newPeerCh:
			// 同步本地的txpool给对端
			go syncTx(peer)
			//先与peer同步状态，然后做总体同步
			syncBlock()
		case <-blockMgr.quit:
			return
		}
	}
}

//块hash是否在本地主链上
func (blockMgr *BlockMgr) checkExistHeaderHash(headerHash *crypto.Hash) (bool, uint64) {
	node := blockMgr.ChainService.Index().LookupNode(headerHash)
	if node == nil {
		return false, 0
	}
	bestNode := blockMgr.ChainService.BestChain().NodeByHeight(node.Height)
	if bestNode == nil {
		return false, 0
	}
	if bytes.Equal(bestNode.Hash.Bytes(), headerHash.Bytes()) {
		return true, bestNode.Height
	}
	return false, 0
}

func (blockMgr *BlockMgr) requestHeaders(peer types.PeerInfoInterface, from, count uint64) error {
	req := types.HeaderReq{FromHeight: from, ToHeight: from + count - 1}
	log.WithField("ip", peer.GetAddr()).Info("req header to")
	return blockMgr.P2pServer.Send(peer.GetMsgRW(), types.MsgTypeHeaderReq, &req)
}

//找到公共祖先
func (blockMgr *BlockMgr) findAncestor(peer types.PeerInfoInterface) (uint64, error) {
	timeout := time.After(time.Second * maxNetworkTimeout)
	remoteHeight := peer.GetHeight()
	fromHeight := blockMgr.ChainService.BestChain().Height()

	//在发出请求的过程中，其他节点的新的块可能已经同步到本地了,因此可以多获取一些
	err := blockMgr.requestHeaders(peer, fromHeight, maxHeaderHashCountReq)
	if err != nil {
		log.WithField("err", err).WithField("fromeheight", fromHeight).Info("req headers ")
		return 0, err
	}

	select {
	case hashs := <-blockMgr.headerHashCh:
		len := len(hashs)
		for i := len - 1; i >= 0; i-- {
			b, h := blockMgr.checkExistHeaderHash(hashs[i].headerHash)
			if b {
				//found it
				return h, nil
			}
		}
		return 0, ErrNoCommonAncesstor
	case <-timeout:
		return 0, ErrFindAncesstorTimeout
	}

	//通过二分查找找相同的HASH
	var ancestor uint64 = 0
	var tmpFrom uint64 = 0
	var tmpEnd uint64 = remoteHeight
	for tmpFrom+1 < tmpEnd {
		timer := time.NewTimer(time.Second * maxNetworkTimeout)
		err = blockMgr.requestHeaders(peer, (tmpFrom+tmpEnd)/2, 1)
		if err != nil {
			return 0, err
		}

		select {
		case hash := <-blockMgr.headerHashCh:
			if len(hash) > 0 {
				b, h := blockMgr.checkExistHeaderHash(hash[0].headerHash)
				if b {
					ancestor = h
					tmpFrom = (tmpFrom + tmpEnd) / 2
				} else {
					tmpEnd = (tmpFrom + tmpEnd) / 2
				}
			} else {
				log.Error("peer response head hash len = 0")
			}

			timer.Stop()
		case <-timer.C:
			return 0, ErrFindAncesstorTimeout
		}
	}
	return ancestor, nil
}

func (blockMgr *BlockMgr) clearSyncCh() {
	select {
	case <-blockMgr.headerHashCh:
	default:
	}

	select {
	case <-blockMgr.blocksCh:
	default:
	}

	select {
	case <-blockMgr.syncTimerCh:
	default:
	}

	blockMgr.allTasks = newHeightSortedMap()

	blockMgr.pendingSyncTasks.Range(func(key, value interface{}) bool {
		blockMgr.pendingSyncTasks.Delete(key)
		return true
	})
}

func (blockMgr *BlockMgr) batchReqBlocks(hashs []crypto.Hash, mapHeightHash map[crypto.Hash]uint64, errCh chan error) {
	req := &types.BlockReq{BlockHashs: hashs}

	var maxHeight uint64 = 0
	var minHeight uint64 = 0xffffffffffffffff
	for _, value := range mapHeightHash {
		if value > maxHeight {
			maxHeight = value
		}
		if value < minHeight {
			minHeight = value
		}
	}

	okPeers := make([]types.PeerInfoInterface, 0, getPeersCount(blockMgr.peersInfo))
	blockMgr.peersInfo.Range(func(key, value interface{}) bool {
		peer := value.(types.PeerInfoInterface)
		if peer.GetHeight() >= maxHeight {
			okPeers = append(okPeers, peer)
		}
		return true
	})

	for _, pi := range okPeers {
		blockMgr.syncMut.Lock()
		//blockMgr.sender.SetIdle(bodyReqPeer.GetMsgRW(), false)
		log.WithField("len", len(hashs)).WithField("from", minHeight).WithField("to:", maxHeight).WithField("destIp", pi.GetAddr()).Info("req block body")
		err := blockMgr.P2pServer.Send(pi.GetMsgRW(), types.MsgTypeBlockReq, req)
		//blockMgr.sender.SetIdle(bodyReqPeer, true)
		blockMgr.syncMut.Unlock()

		if err == nil {
			return
		}
	}

	errCh <- fmt.Errorf("p2p no peers")

	return
}

func (blockMgr *BlockMgr) fetchBlocks(peer types.PeerInfoInterface) error {
	blockMgr.syncBlockEvent.Send(event.SyncBlockEvent{EventType: event.StartSyncBlock})
	defer func() {
		blockMgr.state = event.StopSyncBlock
		blockMgr.syncBlockEvent.Send(event.SyncBlockEvent{EventType: event.StopSyncBlock})
	}()

	if blockMgr.state == event.StartSyncBlock {
		log.Info("have fetch blocks")
		return nil
	}

	blockMgr.state = event.StartSyncBlock

	height := peer.GetHeight()
	blockMgr.clearSyncCh()
	headerRoutineExit := false

	//1 获取公共祖先
	commonAncestor, err := blockMgr.findAncestor(peer)
	if err != nil {
		return err
	}

	log.Info("commonAncestor=", commonAncestor)

	errCh := make(chan error)
	quit := make(chan struct{})
	//2 获取所有需要同步的块的hash;然后通知给获取BODY的协程
	go func() {
		commonAncestor += 1
		timer := time.NewTimer(time.Second * maxNetworkTimeout)

		for height >= commonAncestor {
			select {
			case <-quit:
				log.Info("fetch headers goroutine quit")
				return
			default:
				timer.Reset(time.Second * maxNetworkTimeout)

				blockMgr.syncMut.Lock()
				taskLen := blockMgr.allTasks.Len()
				blockMgr.syncMut.Unlock()
				if taskLen >= pendingTimerCount {
					time.Sleep(maxSyncSleepTime)
					continue
				}

				blockMgr.syncMut.Lock()
				//blockMgr.sender.SetIdle(peer, false)
				err := blockMgr.requestHeaders(peer, commonAncestor, maxHeaderHashCountReq)
				log.WithField("commonAncestor", commonAncestor).Info("req header")
				//blockMgr.sender.SetIdle(peer, true)
				blockMgr.syncMut.Unlock()
				if err != nil {
					errCh <- err
					return
				}

				select {
				//每个headerhash作为一个任务
				case tasks := <-blockMgr.headerHashCh:
					for _, task := range tasks {
						blockMgr.syncMut.Lock()
						blockMgr.allTasks.Put(task)
						blockMgr.syncMut.Unlock()
					}
					commonAncestor += uint64(len(tasks))
					log.WithField("tasks len", blockMgr.allTasks.Len()).WithField("newtasks", len(tasks)).Info("get headers")
				case <-timer.C:
					errCh <- ErrGetHeaderHashTimeout
					return

				}
			}
		}
		log.Info("fetch all headers end ****************************")
		headerRoutineExit = true
	}()

	//请求发出去到时候，设置一个超时定时器
	//请求超时后，把对应到pendingSyncTasks中到任务放入到allTasks，再次被请求
	//请求到块都到了，从pendingSyncTasks中删除对应到任务

	//收到block
	go func() {
		delHash := func(b *types.Block) {
			blockMgr.pendingSyncTasks.Range(func(key, value interface{}) bool {
				blockMgr.syncMut.Lock()
				defer blockMgr.syncMut.Unlock()
				hashs := value.(map[crypto.Hash]uint64)
				timer := key.(*time.Timer)

				if _, ok := hashs[*b.Header.Hash()]; ok {
					delete(hashs, *b.Header.Hash())
					if len(hashs) == 0 {
						//所有到block都到了，停止超时定时器
						//timer.Stop()

						blockMgr.syncTimerCh <- timer
					}

					return false
				}
				return true
			})
		}

		checkExist := func(h crypto.Hash) bool {
			found := false
			blockMgr.pendingSyncTasks.Range(func(key, value interface{}) bool {
				blockMgr.syncMut.Lock()
				defer blockMgr.syncMut.Unlock()
				hashs := value.(map[crypto.Hash]uint64)

				if _, ok := hashs[h]; ok {
					found = true
					return false
				}
				return true
			})

			return found
		}

		for {
			select {
			case blocks := <-blockMgr.blocksCh:
				for _, b := range blocks {

					if !checkExist(*b.Header.Hash()) {
						continue
					}

					//log.WithField("height", b.Header.Height).WithField("blk num", len(blocks)).Info("sync block recv block")
					_, _, err := blockMgr.ChainService.ProcessBlock(b)
					if err != nil {
						switch err {
						case chain.ErrBlockExsist, chain.ErrOrphanBlockExsist:
							//删除块高度对应的任务
							delHash(b)
							continue

						default:
							log.WithField("Reason", err).Error("deal sync block")
							errCh <- err
							return
						}
					} else {
						//删除块高度对应的任务
						delHash(b)
					}
				}

			case <-quit:
				log.Info("fetch blocks routine quit")
				return
			}
		}
	}()

	//3获取对应的body
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				//请求发的太快了，需要等待
				count := 0
				blockMgr.pendingSyncTasks.Range(func(key, value interface{}) bool {
					count++
					if count >= pendingTimerCount {
						return false
					}
					return true
				})
				if count >= pendingTimerCount {
					log.Info("req block body routine wait.......")
					time.Sleep(time.Millisecond * maxSyncSleepTime)
					continue
				}

				blockMgr.syncMut.Lock()
				hashs, headerHashs := blockMgr.allTasks.GetSortedHashs(maxBlockCountReq)
				taskLen := blockMgr.allTasks.Len()
				blockMgr.syncMut.Unlock()
				if len(hashs) == 0 {
					if headerRoutineExit && count == 0 {
						//任务完成，触发退出
						log.WithField("tasks len", taskLen).Info("all block sync ok")
						close(errCh)
						return
					}
					time.Sleep(time.Millisecond * maxSyncSleepTime)
					continue
				}

				reqTimer := time.NewTimer(time.Second * maxNetworkTimeout)
				blockMgr.pendingSyncTasks.Store(reqTimer, headerHashs)

				go func() {
					defer reqTimer.Stop()

					select {
					case <-reqTimer.C:
						//得到hash加入到allTasks
						value, ok := blockMgr.pendingSyncTasks.Load(reqTimer)
						if !ok {
							errCh <- fmt.Errorf("timer not in pending task, exception")
							return
						}
						hashs := value.(map[crypto.Hash]uint64)
						for k, v := range hashs {
							blockMgr.syncMut.Lock()

							k := k
							blockMgr.allTasks.Put(&syncHeaderHash{headerHash: &k, height: v})
							blockMgr.syncMut.Unlock()
						}

						blockMgr.pendingSyncTasks.Delete(reqTimer)
						log.Errorf("req block body time out time：%p\n", reqTimer)

					case timer := <-blockMgr.syncTimerCh:
						//timer.Stop()
						//请求到block都到了，停止此定时器
						blockMgr.pendingSyncTasks.Delete(timer)
					case <-quit:
						log.Info("fetch block timer goroutine quit")
						return
					}
				}()
				blockMgr.batchReqBlocks(hashs, headerHashs, errCh)
			}
		}
	}()

	select {
	case err := <-errCh:
		close(quit)
		return err
	}
}

func (blockMgr *BlockMgr) handlePeerState(peer *types.PeerInfo, peerState *types.PeerState) {
	peer.SetHeight(uint64(peerState.Height))
}

//自己的高度通知出去，对端收到请求后，把自己本地的高度返回
func (blockMgr *BlockMgr) handleReqPeerState(peer *types.PeerInfo, peerState *types.PeerStateReq) {
	peer.SetHeight(uint64(peerState.Height))

	blockMgr.P2pServer.SendAsync(peer.GetMsgRW(), types.MsgTypePeerState, &types.PeerState{
		Height: uint64(blockMgr.ChainService.BestChain().Height()),
	})
}

func (blockMgr *BlockMgr) GetBestPeerInfo() types.PeerInfoInterface {
	var curPeer types.PeerInfoInterface
	blockMgr.peersInfo.Range(func(key, value interface{}) bool {
		pi := value.(types.PeerInfoInterface)
		if curPeer != nil {
			if curPeer.GetHeight() < pi.GetHeight() {
				curPeer = pi
			} else if curPeer.GetHeight() == pi.GetHeight() {
				if rand.Intn(2) == 0 {
					curPeer = pi
				}
			}
		} else {
			curPeer = pi
		}
		return true
	})

	return curPeer
}

func (blockMgr *BlockMgr) checkHeaderChain(chain []types.BlockHeader) error {
	// Do a sanity check that the provided chain is actually ordered and linked
	for i := 1; i < len(chain); i++ {
		if chain[i].Height != chain[i-1].Height+1 || !chain[i].PreviousHash.IsEqual(chain[i-1].Hash()) {
			// Chain broke ancestry, log a message (programming error) and skip insertion
			log.WithField("number", chain[i].Height).
				WithField("hash", hex.EncodeToString(chain[i].Hash().Bytes())).
				WithField("prevnumber", chain[i-1].Height).
				WithField("prevhash", hex.EncodeToString(chain[i-1].Hash().Bytes())).
				WithField("!= parent", hex.EncodeToString(chain[i].PreviousHash.Bytes())).
				Error("Non contiguous header")

			return ErrNotContinueHeader
		}

		for _, blockValidator := range blockMgr.ChainService.BlockValidator() {
			err := blockValidator.VerifyHeader(&chain[i], &chain[i-1])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

package blockmgr

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/drep-project/drep-chain/chain/service/chainservice"
	"math/big"
	"time"

	"github.com/drep-project/dlog"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
)

type tasksTxsSync struct {
	txs  []*chainTypes.Transaction
	peer *chainTypes.PeerInfo
}

//同步缓存池中的交易
func (blockMgr *BlockMgr) syncTxs() {
	for {
		select {
		case task := <-blockMgr.taskTxsCh:
			count := len(task.txs) / maxTxsCount
			var i int
			for i = 0; i < count; i++ {
				blockMgr.P2pServer.Send(task.peer.GetMsgRW(), chainTypes.MsgTypeTransaction, task.txs[i*maxTxsCount:(i+1)*maxTxsCount])
				time.Sleep(time.Millisecond * maxSyncSleepTime)
			}

			blockMgr.P2pServer.Send(task.peer.GetMsgRW(), chainTypes.MsgTypeTransaction, task.txs[count*maxTxsCount:])
		case <-blockMgr.quit:
			return
		}
	}
}

func (blockMgr *BlockMgr) synchronise() {
	timer := time.NewTicker(time.Second * 10)
	defer timer.Stop()

	sync := func() {
		pi := blockMgr.GetBestPeerInfo()
		if pi == nil {
			return
		}
		currentHeight := blockMgr.ChainService.BestChain.Height()
		if pi.GetHeight() > currentHeight {
			fmt.Println("************", pi.GetHeight(), ">", currentHeight)
			err := blockMgr.fetchBlocks(pi)
			if err != nil {
				dlog.Warn("sync block from peer", "err:", err)
			}
		}
	}

	syncTx := func(peer *chainTypes.PeerInfo) {
		//保证能把pending里面的所有tx全部取出来
		txs := blockMgr.transactionPool.GetPending(new(big.Int).SetUint64(0xffffffffffffffff))
		txs2 := blockMgr.transactionPool.GetQueue()

		txs = append(txs, txs2...)
		blockMgr.taskTxsCh <- tasksTxsSync{peer: peer, txs: append(txs, txs2...)}
	}

	for {
		select {
		case <-timer.C:
			sync()
		case peer := <-blockMgr.newPeerCh:
			// 同步本地的txpool给对端
			go syncTx(peer)
			//先与peer同步状态，然后做总体同步
			sync()
		case <-blockMgr.quit:
			return
		}
	}
}

//块hash是否在本地主链上
func (blockMgr *BlockMgr) checkExistHeaderHash(headerHash *crypto.Hash) (bool, uint64) {
	node := blockMgr.ChainService.Index.LookupNode(headerHash)
	if node == nil {
		return false, 0
	}
	bestNode := blockMgr.ChainService.BestChain.NodeByHeight(node.Height)
	if bestNode == nil {
		return false, 0
	}
	if bytes.Equal(bestNode.Hash.Bytes(), headerHash.Bytes()) {
		return true, bestNode.Height
	}
	return false, 0
}

func (blockMgr *BlockMgr) requestHeaders(peer *chainTypes.PeerInfo, from, count uint64) error {
	req := chainTypes.HeaderReq{FromHeight: uint64(from), ToHeight: uint64(from + count - 1)}
	return blockMgr.P2pServer.Send(peer.GetMsgRW(), chainTypes.MsgTypeHeaderReq, &req)
}

//找到公共祖先
func (blockMgr *BlockMgr) findAncestor(peer *chainTypes.PeerInfo) (uint64, error) {
	timeout := time.After(time.Second * maxNetworkTimeout * 4)
	remoteHeight := peer.GetHeight()

	//本地高度和远程高度一样的情况下，对应的HASH是否相同
	fromHeight := blockMgr.ChainService.BestChain.Height()

	//在发出请求的过程中，其他节点的新的块可能已经同步到本地了,因此可以多获取一些
	err := blockMgr.requestHeaders(peer, fromHeight, maxHeaderHashCountReq)
	if err != nil {
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
	case <-timeout:
		return 0, ErrFindAncesstorTimeout
	}

	//通过二分查找找相同的HASH
	var ancestor uint64 = 0
	var tmpFrom uint64 = 0
	var tmpEnd uint64 = remoteHeight
	for tmpFrom+1 < tmpEnd {
		timeout = time.After(time.Second * maxNetworkTimeout)
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
				dlog.Error("peer response head hash len = 0")
			}

		case <-timeout:
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

	blockMgr.allTasks = newHeightSortedMap()
	blockMgr.pendingSyncTasks = make(map[crypto.Hash]uint64)
}

func (blockMgr *BlockMgr) batchReqBlocks(hashs []crypto.Hash, errCh chan error) {
	req := &chainTypes.BlockReq{BlockHashs: hashs}
	successNum := 0

	for _, pi := range blockMgr.peersInfo {
		blockMgr.syncMut.Lock()
		//blockMgr.P2pServer.SetIdle(bodyReqPeer.GetMsgRW(), false)
		err := blockMgr.P2pServer.Send(pi.GetMsgRW(), chainTypes.MsgTypeBlockReq, req)
		//blockMgr.P2pServer.SetIdle(bodyReqPeer, true)
		blockMgr.syncMut.Unlock()

		successNum ++

		if err == nil && successNum >= 2 {
			return
		}
	}

	if successNum == 0 {
		errCh <- fmt.Errorf("p2p no peers")
	}

	return
}

func (blockMgr *BlockMgr) fetchBlocks(peer *chainTypes.PeerInfo) error {
	blockMgr.syncBlockEvent.Send(event.SyncBlockEvent{EventType: event.StartSyncBlock})
	defer blockMgr.syncBlockEvent.Send(event.SyncBlockEvent{EventType: event.StopSyncBlock})

	height := peer.GetHeight()
	blockMgr.clearSyncCh()
	headerRoutineExit := false
	//1 获取公共祖先
	commonAncestor, err := blockMgr.findAncestor(peer)
	if err != nil {
		return err
	}

	errCh := make(chan error)
	quit := make(chan struct{})
	//2 获取所有需要同步的块的hash;然后通知给获取BODY的协程
	go func() {
		commonAncestor += 1
		timer := time.NewTimer(time.Second * maxNetworkTimeout)

		for height >= commonAncestor {
			timer.Reset(time.Second * maxNetworkTimeout)

			if blockMgr.allTasks.Len() >= maxHeaderHashCountReq {
				time.Sleep(time.Millisecond * 100)
				continue
			}

			blockMgr.syncMut.Lock()
			//blockMgr.P2pServer.SetIdle(peer, false)
			err := blockMgr.requestHeaders(peer, commonAncestor, maxHeaderHashCountReq)
			//blockMgr.P2pServer.SetIdle(peer, true)
			blockMgr.syncMut.Unlock()
			if err != nil {
				errCh <- err
				return
			}

			select {
			//每个headhash作为一个任务
			case tasks := <-blockMgr.headerHashCh:
				for _, task := range tasks {
					blockMgr.syncMut.Lock()
					blockMgr.allTasks.Put(task)
					blockMgr.syncMut.Unlock()
				}
				commonAncestor += uint64(len(tasks))
				dlog.Info("fetchBlocks", "tasks len", blockMgr.allTasks.Len())

			case <-timer.C:
				errCh <- ErrGetHeaderHashTimeout
				return
			case <-quit:
				dlog.Info("fetch headers goroutine quit")
				return
			}
		}
		dlog.Info("fetch all headers end ****************************")
		headerRoutineExit = true
	}()

	//3获取对应的body
	go func() {
		for {
			blockMgr.syncMut.Lock()
			hashs, headerHashs := blockMgr.allTasks.GetSortedHashs(maxBlockCountReq)
			blockMgr.syncMut.Unlock()
			if len(hashs) == 0 {
				if headerRoutineExit {
					//任务完成，触发退出
					dlog.Info("all block sync ok", "tasks len", blockMgr.allTasks.Len())
					close(errCh)
					return
				}
				time.Sleep(time.Millisecond * maxSyncSleepTime)
				continue
			}

			blockMgr.pendingSyncTasks = headerHashs

			go blockMgr.batchReqBlocks(hashs, errCh)

			//最多等待一分钟
			timeout := time.After(time.Second * maxNetworkTimeout * 6)
			validCh := make(chan bool)
			// 等待，直到有效的块到来后，才进入到下个循环
			go func() {
				deletedHash := false
				for {
					select {
					case blocks := <-blockMgr.blocksCh:
						for _, b := range blocks {
							if _, ok := blockMgr.pendingSyncTasks[*b.Header.Hash()]; !ok {
								continue
							}
							dlog.Info("sync block recv block", "height", b.Header.Height, "blk num", len(blocks))
							//删除块高度对应的任务
							delete(blockMgr.pendingSyncTasks, *b.Header.Hash())
							deletedHash = true
							_, _, err := blockMgr.ChainService.ProcessBlock(b)
							if err != nil {
								switch err {
								case chainservice.ErrBlockExsist, chainservice.ErrOrphanBlockExsist:
									continue
								default:
									dlog.Error("deal sync block", "err", err)
									errCh <- err
									return
								}
							}
						}

						if deletedHash {
							for k, v := range blockMgr.pendingSyncTasks {
								blockMgr.syncMut.Lock()
								blockMgr.allTasks.Put(&syncHeaderHash{headerHash: &k, height: v})
								blockMgr.syncMut.Unlock()
							}
							validCh <- true
							return
						}
					case <-timeout:
						errCh <- ErrGetBlockTimeout
						return
					case <-quit:
						dlog.Info("fetch blocks routine quit")
						return
					}
				}
			}()

			<-validCh
		}
	}()

	select {
	case err := <-errCh:
		close(quit)
		return err
	}
}

func (blockMgr *BlockMgr) handlePeerState(peer *chainTypes.PeerInfo, peerState *chainTypes.PeerState) {
	peer.SetHeight(uint64(peerState.Height))
}

//自己的高度通知出去，对端收到请求后，把自己本地的高度返回
func (blockMgr *BlockMgr) handleReqPeerState(peer *chainTypes.PeerInfo, peerState *chainTypes.PeerStateReq) {
	peer.SetHeight(uint64(peerState.Height))

	blockMgr.P2pServer.SendAsync(peer.GetMsgRW(), chainTypes.MsgTypePeerState, &chainTypes.PeerState{
		Height: uint64(blockMgr.ChainService.BestChain.Height()),
	})
}

func (blockMgr *BlockMgr) GetBestPeerInfo() *chainTypes.PeerInfo {
	var curPeer *chainTypes.PeerInfo
	for _, pi := range blockMgr.peersInfo {
		if curPeer != nil {
			if curPeer.GetHeight() < pi.GetHeight() {
				curPeer = pi
			}
		} else {
			curPeer = pi
		}
	}
	return curPeer
}

func (blockMgr *BlockMgr) checkHeaderChain(chain []chainTypes.BlockHeader) error {
	// Do a sanity check that the provided chain is actually ordered and linked
	for i := 1; i < len(chain); i++ {
		if chain[i].Height != chain[i-1].Height+1 || !chain[i].PreviousHash.IsEqual(chain[i-1].Hash()) {
			// Chain broke ancestry, log a message (programming error) and skip insertion
			dlog.Error("Non contiguous header", "number", chain[i].Height, "hash", hex.EncodeToString(chain[i].Hash().Bytes()),
				"prevnumber", chain[i-1].Height, "prevhash", hex.EncodeToString(chain[i-1].Hash().Bytes()), "!= parent", hex.EncodeToString(chain[i].PreviousHash.Bytes()))

			return ErrNotContinueHeader
		}

		err := blockMgr.ChainService.BlockValidator.VerifyHeader(&chain[i], &chain[i-1])
		if err != nil {
			return err
		}
	}
	return nil
}

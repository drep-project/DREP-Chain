package service

import (
	"bytes"
	"encoding/hex"
	"fmt"
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
func (chainService *ChainService) syncTxs() {
	for {
		select {
		case task := <-chainService.taskTxsCh:
			count := len(task.txs) / maxTxsCount
			var i int
			for i = 0; i < count; i++ {
				chainService.P2pServer.Send(task.peer.GetMsgRW(), chainTypes.MsgTypeTransaction, task.txs[i*maxTxsCount:(i+1)*maxTxsCount])
				time.Sleep(time.Millisecond * maxSyncSleepTime)
			}

			chainService.P2pServer.Send(task.peer.GetMsgRW(), chainTypes.MsgTypeTransaction, task.txs[count*maxTxsCount:])
		case <-chainService.quit:
			return
		}
	}
}

func (chainService *ChainService) synchronise() {
	timer := time.NewTicker(time.Second * 10)
	defer timer.Stop()

	sync := func() {
		pi := chainService.GetBestPeerInfo()
		if pi == nil {
			return
		}
		if pi.GetHeight() > chainService.BestChain.Height() {
			fmt.Println("************", pi.GetHeight(), ">", chainService.BestChain.Height())
			err := chainService.fetchBlocks(pi)
			if err != nil {
				dlog.Warn("sync block from peer", "err:", err)
			}
		}
	}

	syncTx := func(peer *chainTypes.PeerInfo) {
		//保证能把pending里面的所有tx全部取出来
		txs := chainService.transactionPool.GetPending(new(big.Int).SetUint64(0xffffffffffffffff))
		txs2 := chainService.transactionPool.GetQueue()

		txs = append(txs, txs2...)
		chainService.taskTxsCh <- tasksTxsSync{peer: peer, txs: append(txs, txs2...)}
	}

	for {
		select {
		case <-timer.C:
			sync()
		case peer := <-chainService.newPeerCh:
			// 同步本地的txpool给对端
			go syncTx(peer)
			//先与peer同步状态，然后做总体同步
			sync()
		case <-chainService.quit:
			return
		}
	}
}

//块hash是否在本地主链上
func (chainService *ChainService) checkExistHeaderHash(headerHash *crypto.Hash) (bool, uint64) {
	node := chainService.Index.LookupNode(headerHash)
	if node == nil {
		return false, 0
	}
	bestNode := chainService.BestChain.NodeByHeight(node.Height)
	if bestNode == nil {
		return false, 0
	}
	if bytes.Equal(bestNode.Hash.Bytes(), headerHash.Bytes()) {
		return true, bestNode.Height
	}
	return false, 0
}

func (cs *ChainService) requestHeaders(peer *chainTypes.PeerInfo, from, count uint64) error {
	req := chainTypes.HeaderReq{FromHeight: uint64(from), ToHeight: uint64(from + count - 1)}
	return cs.P2pServer.Send(peer.GetMsgRW(), chainTypes.MsgTypeHeaderReq, &req)
}

//找到公共祖先
func (cs *ChainService) findAncestor(peer *chainTypes.PeerInfo) (uint64, error) {
	timeout := time.After(time.Second * maxNetworkTimeout * 4)
	remoteHeight := peer.GetHeight()

	//本地高度和远程高度一样的情况下，对应的HASH是否相同
	fromHeight := cs.BestChain.Height()

	//在发出请求的过程中，其他节点的新的块可能已经同步到本地了,因此可以多获取一些
	err := cs.requestHeaders(peer, fromHeight, maxHeaderHashCountReq)
	if err != nil {
		return 0, err
	}

	select {
	case hashs := <-cs.headerHashCh:
		len := len(hashs)
		for i := len - 1; i >= 0; i-- {
			b, h := cs.checkExistHeaderHash(hashs[i].headerHash)
			if b {
				//found it
				return h, nil
			}
		}
	case <-timeout:
		return 0, fmt.Errorf("findAncestor timeout")
	}

	//通过二分查找找相同的HASH
	var ancestor uint64 = 0
	var tmpFrom uint64 = 0
	var tmpEnd uint64 = remoteHeight
	for ; tmpFrom+1 < tmpEnd; {
		timeout = time.After(time.Second * maxNetworkTimeout)
		err = cs.requestHeaders(peer, (tmpFrom+tmpEnd)/2, 1)
		if err != nil {
			return 0, err
		}

		select {
		case hash := <-cs.headerHashCh:
			if len(hash) > 0{
				b, h := cs.checkExistHeaderHash(hash[0].headerHash)
				if b {
					ancestor = h
					tmpFrom = (tmpFrom + tmpEnd) / 2
				} else {
					tmpEnd = (tmpFrom + tmpEnd) / 2
				}
			} else{
				dlog.Error("peer response head hash len = 0")
			}

		case <-timeout:
			return 0, fmt.Errorf("findAncestor timeout")
		}
	}
	fmt.Println("common ancenstor:", ancestor)
	return ancestor, nil
}

func (cs *ChainService) clearSyncCh() {
	select {
	case <-cs.headerHashCh:
	default:
	}

	select {
	case <-cs.blocksCh:
	default:
	}

	cs.allTasks = newHeightSortedMap()
	cs.pendingSyncTasks = make(map[crypto.Hash]uint64)
}

func (cs *ChainService) batchReqBlocks(hashs []crypto.Hash) {
	for {
		if len(cs.peersInfo) == 0 {
			//errCh <- fmt.Errorf("no idle peers, sync block stop")
			time.Sleep(time.Millisecond * maxSyncSleepTime)
			continue
		}
		req := &chainTypes.BlockReq{BlockHashs: hashs}
		errNum := 0
		for _, pi := range cs.peersInfo {
			cs.syncMut.Lock()
			//cs.P2pServer.SetIdle(bodyReqPeer.GetMsgRW(), false)
			err := cs.P2pServer.Send(pi.GetMsgRW(), chainTypes.MsgTypeBlockReq, req)
			//cs.P2pServer.SetIdle(bodyReqPeer, true)
			cs.syncMut.Unlock()
			if err != nil {
				errNum ++
			}
		}
		if errNum < len(cs.peersInfo) {
			break
		}
	}
}

func (cs *ChainService) fetchBlocks(peer *chainTypes.PeerInfo) error {
	cs.syncBlockEvent.Send(event.SyncBlockEvent{EventType: event.StartSyncBlock})
	defer cs.syncBlockEvent.Send(event.SyncBlockEvent{EventType: event.StopSyncBlock})

	height := peer.GetHeight()
	cs.clearSyncCh()
	headerRoutineExit := false
	//1 获取公共祖先
	commonAncestor, err := cs.findAncestor(peer)
	if err != nil {
		return err
	}

	errCh := make(chan error)
	quit := make(chan struct{})
	//2 获取所有需要同步的块的hash;然后通知给获取BODY的协程
	go func() {
		commonAncestor += 1
		for height > commonAncestor {
			timeout := time.After(time.Second * maxNetworkTimeout)

			cs.syncMut.Lock()
			//cs.P2pServer.SetIdle(peer, false)
			err := cs.requestHeaders(peer, commonAncestor, maxHeaderHashCountReq)
			//cs.P2pServer.SetIdle(peer, true)
			cs.syncMut.Unlock()
			if err != nil {
				errCh <- err
				return
			}

			select {
			//任务从远端传来
			case tasks := <-cs.headerHashCh:
				for _, task := range tasks {
					cs.syncMut.Lock()
					cs.allTasks.Put(task)
					cs.syncMut.Unlock()
				}
				commonAncestor += uint64(len(tasks))
				dlog.Info("fetchBlocks", "tasks len", cs.allTasks.Len())

			case <-timeout:
				errCh <- fmt.Errorf("get header hash timeout")
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
			cs.syncMut.Lock()
			hashs, headerHashs := cs.allTasks.GetSortedHashs(maxBlockCountReq)
			cs.syncMut.Unlock()
			if len(hashs) == 0 {
				if headerRoutineExit {
					//任务完成，触发退出
					dlog.Info("all block sync ok", "tasks len", cs.allTasks.Len())
					close(errCh)
					return
				}
				time.Sleep(time.Millisecond * maxSyncSleepTime)
				continue
			}

			cs.pendingSyncTasks = headerHashs
			go cs.batchReqBlocks(hashs)

			//最多等待一分钟
			timeout := time.After(time.Second * maxNetworkTimeout * 12)

			select {
			case blocks := <-cs.blocksCh:
				for _, b := range blocks {
					dlog.Info("sync block recv block", "height", b.Header.Height)

					//删除块高度对应的任务
					delete(cs.pendingSyncTasks, *b.Header.Hash())

					_, _, err := cs.ProcessBlock(b)
					//dlog.Info("sync block recv block","height", b.Header.Height, "process result", err)
					if err != nil && err != errBlockExsist && err != errOrphanBlockExsist {
						dlog.Error("deal sync block", "err", err)
						errCh <- err
						return
					}
				}

				for k, v := range cs.pendingSyncTasks {
					cs.syncMut.Lock()
					cs.allTasks.Put(&syncHeaderHash{headerHash: &k, height: v})
					cs.syncMut.Unlock()
				}

			case <-timeout:
				errCh <- fmt.Errorf("fetch blocks timeout")
			case <-quit:
				dlog.Info("fetch blocks routine quit")
				return
			}
		}
	}()

	select {
	case err := <-errCh:
		close(quit)
		return err
	}
}

func (chainService *ChainService) handlePeerState(peer *chainTypes.PeerInfo, peerState *chainTypes.PeerState) {
	peer.SetHeight(uint64(peerState.Height))
}

//自己的高度通知出去，对端收到请求后，把自己本地的高度返回
func (chainService *ChainService) handleReqPeerState(peer *chainTypes.PeerInfo, peerState *chainTypes.PeerStateReq) {
	peer.SetHeight(uint64(peerState.Height))

	chainService.P2pServer.SendAsync(peer.GetMsgRW(), chainTypes.MsgTypePeerState, &chainTypes.PeerState{
		Height: uint64(chainService.BestChain.Height()),
	})
}

func (chainService *ChainService) GetBestPeerInfo() *chainTypes.PeerInfo {
	var curPeer *chainTypes.PeerInfo
	for _, pi := range chainService.peersInfo {
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

func (cs *ChainService) checkHeaderChain(chain []chainTypes.BlockHeader) (error) {
	// Do a sanity check that the provided chain is actually ordered and linked
	for i := 1; i < len(chain); i++ {
		if chain[i].Height != chain[i-1].Height+1 || !chain[i].PreviousHash.IsEqual(chain[i-1].Hash()) {
			// Chain broke ancestry, log a message (programming error) and skip insertion
			dlog.Error("Non contiguous header", "number", chain[i].Height, "hash", hex.EncodeToString(chain[i].Hash().Bytes()),
				"prevnumber", chain[i-1].Height, "prevhash", hex.EncodeToString(chain[i-1].Hash().Bytes()), "!= parent", hex.EncodeToString(chain[i].PreviousHash.Bytes()))

			return fmt.Errorf("non contiguous headers")
		}

		err := cs.VerifyHeader(&chain[i],&chain[i-1])
		if err != nil {
			return  err
		}
	}
	return nil
}

func (cs *ChainService) deriveMerkleRoot(txs []*chainTypes.Transaction) []byte {
	if len(txs) == 0{
		return []byte{}
	}
	txHashes, _ := cs.GetTxHashes(txs)
	merkle := cs.DatabaseService.NewMerkle(txHashes)
	return merkle.Root.Hash
}

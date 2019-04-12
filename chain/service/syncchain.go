package service

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/drep-project/dlog"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	p2pTypes "github.com/drep-project/drep-chain/network/types"
	"time"
)

func (chainService *ChainService) synchronise() {
	for {
		chainService.P2pServer.Broadcast(&chainTypes.ReqPeerState{
			Height: chainService.BestChain.Height(),
		})
		time.Sleep(time.Second * 5)

		peer, state := chainService.GetBestPeer()
		if peer == nil || state == nil || state.Height < 1 {
			continue
		}

		if state.Height > chainService.BestChain.Height() {
			err := chainService.fetchBlocks(peer, state.Height)
			if err != nil {
				dlog.Warn("sync block from peer", "err:", err)
			}
		}
	}
}

//块hash是否在本地主链上
func (chainService *ChainService) checkExistHeaderHash(headerHash *crypto.Hash) (bool, int64) {
	node := chainService.Index.LookupNode(headerHash)
	if node == nil {
		return false, -1
	}
	bestNode := chainService.BestChain.NodeByHeight(node.Height)
	if bestNode == nil {
		return false, -1
	}
	if bytes.Equal(bestNode.Hash.Bytes(), headerHash.Bytes()) {
		return true, bestNode.Height
	}
	return false, -1
}

func (cs *ChainService) requestHeaders(peer *p2pTypes.Peer, from, count int64) error {
	req := chainTypes.HeaderReq{FromHeight: from, ToHeight: from + count - 1}
	return cs.P2pServer.Send(peer, &req)
}

//找到公共祖先
func (cs *ChainService) findAncestor(peer *p2pTypes.Peer, remoteHeight int64) (int64, error) {
	timeout := time.After(time.Second * maxNetworkTimeout)

	//本地高度和远程高度一样的情况下，对应的HASH是否相同
	fromHeight := cs.BestChain.Height()

	//在发出请求的过程中，其他节点的新的块可能已经同步到本地了,因此可以多获取一些
	err := cs.requestHeaders(peer, fromHeight, maxHeaderHashCountReq)
	if err != nil {
		return -1, err
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
		return -1, fmt.Errorf("findAncestor timeout")
	}

	//通过二分查找找相同的HASH
	var ancestor int64 = 0
	var tmpFrom int64 = 0
	var tmpEnd int64 = remoteHeight
	for ; tmpFrom+1 < tmpEnd; {
		timeout = time.After(time.Second * maxNetworkTimeout)
		err = cs.requestHeaders(peer, (tmpFrom+tmpEnd)/2, 1)
		if err != nil {
			return -1, err
		}

		select {
		case hash := <-cs.headerHashCh:
			b, h := cs.checkExistHeaderHash(hash[0].headerHash)
			if b {
				ancestor = h
				tmpFrom = (tmpFrom + tmpEnd) / 2
			} else {
				tmpEnd = (tmpFrom + tmpEnd) / 2
			}
		case <-timeout:
			return -1, fmt.Errorf("findAncestor timeout")
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
	cs.pendingSyncTasks = make(map[crypto.Hash]int64)
}

func (cs *ChainService) bathReqBlocks(hashs []crypto.Hash) {
	for {
		peers := cs.P2pServer.GetIdlePeers(maxPeerCountReq)
		if len(peers) == 0 {
			//errCh <- fmt.Errorf("no idle peers, sync block stop")
			time.Sleep(time.Millisecond * maxSyncSleepTime)
			continue
		}
		req := &chainTypes.BlockReq{BlockHashs: hashs}
		errNum := 0
		for _, bodyReqPeer := range peers {
			cs.syncMut.Lock()
			cs.P2pServer.SetIdle(bodyReqPeer, false)
			err := cs.P2pServer.Send(bodyReqPeer, req)
			cs.P2pServer.SetIdle(bodyReqPeer, true)
			cs.syncMut.Unlock()
			if err != nil {
				errNum ++
			}
		}
		if errNum < len(peers) {
			break
		}
	}
}

func (cs *ChainService) fetchBlocks(peer *p2pTypes.Peer, height int64) error {
	cs.syncBlockEvent.Send(event.SyncBlockEvent{EventType: event.StartSyncBlock})
	defer cs.syncBlockEvent.Send(event.SyncBlockEvent{EventType: event.StopSyncBlock})

	cs.clearSyncCh()
	headerRoutineExit := false
	//1 获取公共祖先
	commonAncestor, err := cs.findAncestor(peer, height)
	if err != nil {
		return err
	}

	errCh := make(chan error)
	quit := make(chan struct{})
	//2 获取所有需要同步的块的hash;然后通知给获取BODY的协程
	go func() {
		commonAncestor += 1
		for height >= commonAncestor {
			timeout := time.After(time.Second * maxNetworkTimeout)

			cs.syncMut.Lock()
			cs.P2pServer.SetIdle(peer, false)
			err := cs.requestHeaders(peer, commonAncestor, maxHeaderHashCountReq)
			cs.P2pServer.SetIdle(peer, true)
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
				commonAncestor += int64(len(tasks))
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
			go cs.bathReqBlocks(hashs)

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

func (chainService *ChainService) handlePeerState(peer *p2pTypes.Peer, peerState *chainTypes.PeerState) {
	chainService.peerStateLock.Lock()
	defer chainService.peerStateLock.Unlock()
	//get bestpeers
	if _, ok := chainService.peerStateMap[string(peer.Ip)]; ok {
		chainService.peerStateMap[string(peer.Ip)].Height = peerState.Height
	} else {
		chainService.peerStateMap[string(peer.Ip)] = peerState
	}
}

func (chainService *ChainService) handleReqPeerState(peer *p2pTypes.Peer, peerState *chainTypes.ReqPeerState) {
	chainService.peerStateLock.Lock()
	defer chainService.peerStateLock.Unlock()
	if _, ok := chainService.peerStateMap[string(peer.Ip)]; ok {
		chainService.peerStateMap[string(peer.Ip)].Height = peerState.Height
	} else {
		chainService.peerStateMap[string(peer.Ip)] = &chainTypes.PeerState{Height: peerState.Height}
	}

	chainService.P2pServer.SendAsync(peer, &chainTypes.PeerState{
		Height: chainService.BestChain.Height(),
	})
}

func (chainService *ChainService) GetBestPeer() (*p2pTypes.Peer, *chainTypes.PeerState) {
	peers := chainService.P2pServer.Peers()
	if len(peers) == 0 {
		return nil, nil
	}
	curPeer := peers[0]
	chainService.peerStateLock.Lock()
	defer chainService.peerStateLock.Unlock()
	for i := 1; i < len(peers); i++ {
		peerId := string(peers[i].Ip)
		curPeerId := string(curPeer.Ip)
		if _, ok := chainService.peerStateMap[peerId]; !ok {
			chainService.peerStateMap[peerId] = &chainTypes.PeerState{Height: 0}
		}
		if _, ok := chainService.peerStateMap[curPeerId]; !ok {
			chainService.peerStateMap[curPeerId] = &chainTypes.PeerState{Height: 0}
		}
		if chainService.peerStateMap[peerId].Height > chainService.peerStateMap[curPeerId].Height {
			curPeer = peers[i]
		}
	}
	return curPeer, chainService.peerStateMap[string(curPeer.Ip)]
}

func (cs *ChainService) checkHeaderChain(chain []chainTypes.BlockHeader) (error) {
	// Do a sanity check that the provided chain is actually ordered and linked
	for i := 1; i < len(chain); i++ {
		if chain[i].Height != chain[i-1].Height+1 || chain[i].PreviousHash != *chain[i-1].Hash() {
			// Chain broke ancestry, log a message (programming error) and skip insertion
			dlog.Error("Non contiguous header", "number", chain[i].Height, "hash", hex.EncodeToString(chain[i].Hash().Bytes()),
				"parent", hex.EncodeToString(chain[i].PreviousHash.Bytes()), "prevnumber", chain[i-1].Height, "prevhash", hex.EncodeToString(chain[i-1].Hash().Bytes()))

			return fmt.Errorf("non contiguous headers: item-1:%d  height:%d hash:%s, item:%d height:%d hash:%s",
				i-1, chain[i-1].Height, chain[i-1].Hash().Bytes()[:4], i, chain[i].Height, chain[i].Hash().Bytes()[:4])
		}

		cs.checkHeader(&chain[i])
	}

	return nil
}

func (cs *ChainService) deriveMerkleRoot(txs []*chainTypes.Transaction) []byte {
	txHashes, _ := cs.GetTxHashes(txs)
	merkle := cs.DatabaseService.NewMerkle(txHashes)
	return merkle.Root.Hash
}
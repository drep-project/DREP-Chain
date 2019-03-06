package service

import (
	"fmt"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/crypto"
	p2pTypes "github.com/drep-project/drep-chain/network/types"
	"time"
)

func (chainService *ChainService) fetchBlocks() {
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
			chainService.syncMaxHeightMut.Lock()
			if chainService.syncingMaxHeight == -1 {
				chainService.syncingMaxHeight = state.Height
				fmt.Println(chainService.syncingMaxHeight)
				chainService.syncBlockEvent.Send(event.SyncBlockEvent{EventType: event.StartSyncBlock})
			} else if chainService.syncingMaxHeight < state.Height {
				chainService.syncingMaxHeight = state.Height
			}
			chainService.syncMaxHeightMut.Unlock()

			req := &chainTypes.BlockReq{
				StartHash: *chainService.BestChain.Tip().Hash,
				StopHash: crypto.Hash{},
			}
			chainService.P2pServer.Send(peer, req)
		}
	}
}

func (chainService *ChainService) handlePeerState(peer *p2pTypes.Peer, peerState *chainTypes.PeerState) {
	//get bestpeers
	if _, ok := chainService.peerStateMap[string(peer.Ip)]; ok {
		chainService.peerStateMap[string(peer.Ip)].Height = peerState.Height
	} else {
		chainService.peerStateMap[string(peer.Ip)] = peerState
	}
}

func (chainService *ChainService) handleReqPeerState(peer *p2pTypes.Peer, peerState *chainTypes.ReqPeerState) {

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

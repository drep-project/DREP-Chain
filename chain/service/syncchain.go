package service

import (
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	p2pTypes "github.com/drep-project/drep-chain/network/types"
	"time"
)

func (chainService *ChainService) fetchBlocks() {
	go func() {
		for {
			chainService.P2pServer.Broadcast(&chainTypes.ReqPeerState{
				Height: chainService.DatabaseService.GetMaxHeight(),
			})
			time.Sleep(time.Second * 5)
			peer, state := chainService.GetBestPeer()
			if peer == nil || state == nil || state.Height < 1 {
				continue
			}
			if state.Height > chainService.DatabaseService.GetMaxHeight() {
				req := &chainTypes.BlockReq{Height: chainService.DatabaseService.GetMaxHeight()}
				chainService.P2pServer.Send(peer, req)
			}
		}
	}()
}

func (chainService *ChainService) handlePeerState(peer *p2pTypes.Peer, peerState *chainTypes.PeerState) {
	//get bestpeers
	if _, ok := chainService.peerStateMap[string(peer.PubKey.Serialize())]; ok {
		chainService.peerStateMap[string(peer.PubKey.Serialize())].Height = peerState.Height
	} else {
		chainService.peerStateMap[string(peer.PubKey.Serialize())] = peerState
	}
}

func (chainService *ChainService) handleReqPeerState(peer *p2pTypes.Peer, peerState *chainTypes.ReqPeerState) {

	if _, ok := chainService.peerStateMap[string(peer.PubKey.Serialize())]; ok {
		chainService.peerStateMap[string(peer.PubKey.Serialize())].Height = peerState.Height
	} else {
		chainService.peerStateMap[string(peer.PubKey.Serialize())] = &chainTypes.PeerState{Height: peerState.Height}
	}

	chainService.P2pServer.SendAsync(peer, &chainTypes.PeerState{
		Height: chainService.DatabaseService.GetMaxHeight(),
	})
}

func (chainService *ChainService) GetBestPeer() (*p2pTypes.Peer, *chainTypes.PeerState) {
	peers := chainService.P2pServer.Peers()
	if len(peers) == 0 {
		return nil, nil
	}
	curPeer := peers[0]

	for i := 1; i < len(peers); i++ {
		peerId := string(peers[i].PubKey.Serialize())
		curPeerId := string(curPeer.PubKey.Serialize())
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
	return curPeer, chainService.peerStateMap[string(curPeer.PubKey.Serialize())]
}

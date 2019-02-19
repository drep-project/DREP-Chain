package service

import (
	p2pTypes "github.com/drep-project/drep-chain/network/types"
	"time"
)

func (chainService *ChainService) fetchBlocks() {
	go func() {
		for{
			chainService.p2pServer.Broadcast(&p2pTypes.ReqPeerState{
				//Height:database.GetMaxHeight(),
			})
			time.Sleep(time.Second*5)
			peer := chainService.p2pServer.GetBestPeer()
			if peer == nil || peer.State == nil ||peer.State.Height<1{
				continue
			}
			//if peer.State.Height > database.GetMaxHeight() {
			//	req := &bean.BlockReq{Height:database.GetMaxHeight(), Pk: (*secp256k1.PublicKey)(&n.prvKey.PublicKey)}
			//	n.p2pServer.Send(peer,req)
			//}
		}
	}()
}

func (n *ChainService) handlePeerState(peer *p2pTypes.Peer, peerState *p2pTypes.PeerState) {
	//get bestpeers
	peer.State.Height = peerState.Height
}

func (n *ChainService) handleReqPeerState(peer *p2pTypes.Peer, peerState *p2pTypes.ReqPeerState) {
	peer.State.Height = peerState.Height
	n.p2pServer.SendAsync(peer, &p2pTypes.PeerState{
	//	Height:database.GetMaxHeight(),
	})
}
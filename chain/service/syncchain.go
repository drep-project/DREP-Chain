package service

import (
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	p2pTypes "github.com/drep-project/drep-chain/network/types"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"time"
)

func (chainService *ChainService) fetchBlocks() {
	go func() {
		for{
			chainService.P2pServer.Broadcast(&p2pTypes.ReqPeerState{
				Height:chainService.DatabaseService.GetMaxHeight(),
			})
			time.Sleep(time.Second*5)
			peer := chainService.P2pServer.GetBestPeer()
			if peer == nil || peer.State == nil ||peer.State.Height<1{
				continue
			}
			if peer.State.Height > chainService.DatabaseService.GetMaxHeight() {
				req := &chainTypes.BlockReq{Height:chainService.DatabaseService.GetMaxHeight(), Pk: (*secp256k1.PublicKey)(&chainService.prvKey.PublicKey)}
				chainService.P2pServer.Send(peer,req)
			}
		}
	}()
}

func (chainService *ChainService) handlePeerState(peer *p2pTypes.Peer, peerState *p2pTypes.PeerState) {
	//get bestpeers
	peer.State.Height = peerState.Height
}

func (chainService *ChainService) handleReqPeerState(peer *p2pTypes.Peer, peerState *p2pTypes.ReqPeerState) {
	peer.State.Height = peerState.Height
	chainService.P2pServer.SendAsync(peer, &p2pTypes.PeerState{
		Height : chainService.DatabaseService.GetMaxHeight(),
	})
}
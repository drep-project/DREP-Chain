package service

import (
	"github.com/drep-project/drep-chain/network/p2p"
)

type P2PApi struct {
	p2pService P2P
}

func (p2pApis *P2PApi) GetPeers() []*p2p.Peer{
	return p2pApis.p2pService.Peers()
}

func (p2pApis *P2PApi) AddPeers(addr string) {
	p2pApis.p2pService.AddPeer(addr)
}

func (p2pApis *P2PApi) RemovePeers(addr string) {
	p2pApis.p2pService.RemovePeer(addr)
}

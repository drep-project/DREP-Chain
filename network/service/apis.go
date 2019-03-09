package service

import "github.com/drep-project/drep-chain/network/types"

type P2PApi struct {
	p2pService P2P
}

func (p2pApis *P2PApi) GetPeers() []*types.Peer{
	return p2pApis.p2pService.Peers()
}

func (p2pApis *P2PApi) AddPeers(addr string) {
	p2pApis.p2pService.AddPeer(addr)
}

func (p2pApis *P2PApi) RemovePeers(addr string) {
	p2pApis.p2pService.RemovePeer(addr)
}

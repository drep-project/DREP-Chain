package service

import (
	"github.com/drep-project/DREP-Chain/app"
	"github.com/drep-project/DREP-Chain/network/p2p"
	"github.com/drep-project/DREP-Chain/network/p2p/enode"
)

type P2P interface {
	app.Service
	SendAsync(w p2p.MsgWriter, msgType uint64, msg interface{}) chan error
	Send(w p2p.MsgWriter, msgType uint64, msg interface{}) error
	Peers() []*p2p.Peer
	AddPeer(nodeUrl string) error
	RemovePeer(url string)
	AddProtocols(protocols []p2p.Protocol)
	LocalNode() *enode.Node
	//SubscribeEvents(ch chan *p2p.PeerEvent) event.Subscription
}

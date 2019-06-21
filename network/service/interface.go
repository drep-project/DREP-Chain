package service

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/network/p2p"
)

type P2P interface {
	app.Service
	SendAsync(w p2p.MsgWriter, msgType uint64, msg interface{}) chan error
	Send(w p2p.MsgWriter, msgType uint64, msg interface{}) error
	Peers() []*p2p.Peer
	AddPeer(nodeUrl string) error
	RemovePeer(url string)
	AddProtocols(protocols []p2p.Protocol)
}

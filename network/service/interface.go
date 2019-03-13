package service

import (
	"github.com/drep-project/drep-chain/app"
	p2pTypes "github.com/drep-project/drep-chain/network/types"
)

type P2P interface {
	app.Service

	SendAsync(peer *p2pTypes.Peer, msg interface{}) chan error
	Send(peer *p2pTypes.Peer, msg interface{}) error
	Broadcast(msg interface{})
	Peers() []*p2pTypes.Peer
	GetPeer(ip string) *p2pTypes.Peer
	GetRouter() *p2pTypes.MessageRouter
	AddPeer(addr string) error
	RemovePeer(addr string)
	SetIdle(peer *p2pTypes.Peer, idle bool)
	GetIdlePeers(count int)[]*p2pTypes.Peer
}
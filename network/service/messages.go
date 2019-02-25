package service

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	p2pTypes "github.com/drep-project/drep-chain/network/types"
)
func (p2pService *P2pService) Receive(context actor.Context) {
	routeMsg, ok := context.Message().(*p2pTypes.RouteIn)
	if !ok {
		return
	}
	switch msg := routeMsg.Detail.(type) {
	case *p2pTypes.Ping:
		p2pService.handPing(routeMsg.Peer, msg)
	case *p2pTypes.Pong:
		p2pService.handPong(routeMsg.Peer, msg)
	}
}
package service

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	consensusTypes "github.com/drep-project/drep-chain/pkgs/consensus/types"
	"github.com/drep-project/dlog"
	p2pTypes "github.com/drep-project/drep-chain/network/types"
)

func (consensusService *ConsensusService) Receive(context actor.Context) {
	routeMsg, ok := context.Message().(*p2pTypes.RouteIn)
	if !ok {
		return
	}

	switch msg := routeMsg.Detail.(type) {
	case *consensusTypes.Setup:
		dlog.Debug("receive setup msg "+ routeMsg.Peer.GetAddr())
		consensusService.member.OnSetUp(routeMsg.Peer, msg)
	case *consensusTypes.Commitment:
		dlog.Debug("receive Commitment msg " + routeMsg.Peer.GetAddr())
		consensusService.leader.OnCommit(routeMsg.Peer, msg)
	case *consensusTypes.Challenge:
		dlog.Debug("receive Challenge msg "+ routeMsg.Peer.GetAddr())
		consensusService.member.OnChallenge(routeMsg.Peer, msg)
	case *consensusTypes.Response:
		dlog.Debug("receive Response msg "+ routeMsg.Peer.GetAddr())
		consensusService.leader.OnResponse(routeMsg.Peer, msg)
	case *consensusTypes.Fail:
		dlog.Debug("receive Fail msg")
		consensusService.member.OnFail(routeMsg.Peer, msg)
	}
}
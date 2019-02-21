package service

import (
	"github.com/drep-project/drep-chain/log"
	"github.com/AsynkronIT/protoactor-go/actor"
	consensusTypes "github.com/drep-project/drep-chain/consensus/types"
	p2pTypes "github.com/drep-project/drep-chain/network/types"
)

func (consensusService *ConsensusService) Receive(context actor.Context) {
	routeMsg, ok := context.Message().(*p2pTypes.RouteIn)
	if !ok {
		return
	}
	switch msg := routeMsg.Detail.(type) {
	case *consensusTypes.Setup:
		log.Debug("receive setup msg "+ routeMsg.Peer.GetAddr())
		consensusService.member.OnSetUp(routeMsg.Peer, msg)
	case *consensusTypes.Commitment:
		log.Debug("receive Commitment msg " + routeMsg.Peer.GetAddr())
		consensusService.leader.OnCommit(routeMsg.Peer, msg)
	case *consensusTypes.Challenge:
		log.Debug("receive Challenge msg "+ routeMsg.Peer.GetAddr())
		consensusService.member.OnChallenge(routeMsg.Peer, msg)
	case *consensusTypes.Response:
		log.Debug("receive Response msg "+ routeMsg.Peer.GetAddr())
		consensusService.leader.OnResponse(routeMsg.Peer, msg)
	case *consensusTypes.Fail:
		log.Debug("receive Fail msg")
		consensusService.member.OnFail(routeMsg.Peer, msg)
	}
}
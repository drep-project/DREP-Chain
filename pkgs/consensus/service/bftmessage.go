package service

import (
	"fmt"

	"github.com/drep-project/drep-chain/network/p2p"
	consensusTypes "github.com/drep-project/drep-chain/pkgs/consensus/types"
)

func (cs *ConsensusService) receiveMsg(peer *consensusTypes.PeerInfo, rw p2p.MsgReadWriter) error {
	fmt.Println("ConsensusService peeraddr:", peer.IP())
	for {
		msg, err := rw.ReadMsg()
		if err != nil {
			log.WithField("Reason", err).Info("consensus receive msg")
			return err
		}

		if msg.Size > consensusTypes.MaxMsgSize {
			return ErrMsgSize
		}

		switch msg.Code {
		case consensusTypes.MsgTypeSetUp:
			log.WithField("Msg", "MsgTypeSetUp").WithField("IP", peer.IP()).Debug("Receive Leader message")
			var req consensusTypes.Setup
			if err := msg.Decode(&req); err != nil {
				return fmt.Errorf("setup msg:%v err:%v", msg, err)
			}
			cs.member.OnSetUp(peer, &req)
		case consensusTypes.MsgTypeCommitment:
			log.WithField("Msg", "MsgTypeCommitment").WithField("IP", peer.IP()).Debug("Receive member message")
			var req consensusTypes.Commitment
			if err := msg.Decode(&req); err != nil {
				return fmt.Errorf("commit msg:%v err:%v", msg, err)
			}
			cs.leader.OnCommit(peer, &req)
		case consensusTypes.MsgTypeResponse:
			log.WithField("Msg", "MsgTypeResponse").WithField("IP", peer.IP()).Debug("Receive member message")
			var req consensusTypes.Response
			if err := msg.Decode(&req); err != nil {
				return fmt.Errorf("response msg:%v err:%v", msg, err)
			}
			cs.leader.OnResponse(peer, &req)
		case consensusTypes.MsgTypeChallenge:
			log.WithField("Msg", "MsgTypeChallenge").WithField("IP", peer.IP()).Debug("Receive leader message")
			var req consensusTypes.Challenge
			if err := msg.Decode(&req); err != nil {
				return fmt.Errorf("challenge msg:%v err:%v", msg, err)
			}
			cs.member.OnChallenge(peer, &req)
		case consensusTypes.MsgTypeFail:
			log.WithField("Msg", "MsgTypeFail").WithField("IP", peer.IP()).Debug("Receive leader message")
			var req consensusTypes.Fail
			if err := msg.Decode(&req); err != nil {
				return fmt.Errorf("challenge msg:%v err:%v", msg, err)
			}
			cs.member.OnFail(peer, &req)
		default:
			return fmt.Errorf("consensus unkonw msg type:%d", msg.Code)
		}
	}

	return nil
}

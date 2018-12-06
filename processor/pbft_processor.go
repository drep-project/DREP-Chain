package processor

import (
    "BlockChainTest/bean"
    "BlockChainTest/network"
    "BlockChainTest/pool"
)


type SetupMsg struct {
    Msg *bean.Setup
    Peer *network.Peer
}

type CommitmentMsg struct {
    Msg *bean.Commitment
    Peer *network.Peer
}

type ChallengeMsg struct {
    Msg *bean.Challenge
    Peer *network.Peer
}

type ResponseMsg struct {
    Msg *bean.Response
    Peer *network.Peer
}

type SetUpProcessor struct {
}

func (p *SetUpProcessor) process(peer *network.Peer, msg interface{}) {
    if setUp, ok := msg.(*bean.Setup); ok {
        //if member := store.GetItSelfOnMember(); member != nil {
        //    if !member.ProcessSetUp(setUp) {
        //        log.Println("SOS Help ", *setUp)
        //        store.SetRemainingSetup(setUp)
        //    }
        //} else {
        //    store.SetRemainingSetup(setUp)
        //}
        pool.Push(setUp)
    }
}

type CommitProcessor struct {}

func (p *CommitProcessor) process(peer *network.Peer, msg interface{}) {
    if commitment, ok := msg.(*bean.Commitment); ok {
        //if leader := store.GetItSelfOnLeader(); leader != nil {
        //    leader.ProcessCommit(commitment)
        //}
        pool.Push(&CommitmentMsg{Msg: commitment, Peer:peer})
    }
}

type ChallengeProcessor struct {}

func (p *ChallengeProcessor) process(peer *network.Peer, msg interface{}) {
    if challenge, ok := msg.(*bean.Challenge); ok {
        //if member := store.GetItSelfOnMember(); member != nil {
        //    member.ProcessChallenge(challenge)
        //}
        pool.Push(challenge)
    }
}

type ResponseProcessor struct {}

func (p *ResponseProcessor) process(peer *network.Peer, msg interface{}) {
    if response, ok := msg.(*bean.Response); ok {
        //if leader := store.GetItSelfOnLeader(); leader != nil {
        //    leader.ProcessResponse(response)
        //}
        pool.Push(response)
    }
}
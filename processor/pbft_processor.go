package processor

import (
    "BlockChainTest/bean"
    "BlockChainTest/store"
    "BlockChainTest/network"
    "BlockChainTest/pool"
)

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
        pool.Push(commitment)
    }
}

type ChallengeProcessor struct {}

func (p *ChallengeProcessor) process(peer *network.Peer, msg interface{}) {
    if challenge, ok := msg.(*bean.Challenge); ok {
        if member := store.GetItSelfOnMember(); member != nil {
            member.ProcessChallenge(challenge)
        }
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
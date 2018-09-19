package processor

import (
    "BlockChainTest/bean"
    "BlockChainTest/store"
    "BlockChainTest/log"
)

type SetUpProcessor struct {
}

func (p *SetUpProcessor) process(msg interface{}) {
    if setUp, ok := msg.(*bean.Setup); ok {
        if member := store.GetItSelfOnMember(); member != nil {
            if !member.ProcessSetUp(setUp) {
                log.Println("SOS Help ", *setUp)
                store.SetRemainingSetup(setUp)
            }
        } else {
            store.SetRemainingSetup(setUp)
        }
    }
}

type CommitProcessor struct {
}

func (p *CommitProcessor) process(msg interface{}) {
    if commitment, ok := msg.(*bean.Commitment); ok {
        if leader := store.GetItSelfOnLeader(); leader != nil {
            leader.ProcessCommit(commitment)
        }
    }
}

type ChallengeProcessor struct {
}

func (p *ChallengeProcessor) process(msg interface{}) {
    if challenge, ok := msg.(*bean.Challenge); ok {
        if member := store.GetItSelfOnMember(); member != nil {
            member.ProcessChallenge(challenge)
        }
    }
}

type ResponseProcessor struct {
}

func (p *ResponseProcessor) process(msg interface{}) {
    if response, ok := msg.(*bean.Response); ok {
        if leader := store.GetItSelfOnLeader(); leader != nil {
            leader.ProcessResponse(response)
        }
    }
}
package processor

import (
    "BlockChainTest/bean"
    "BlockChainTest/store"
)

type SetUpProcessor struct {
}

func (p *SetUpProcessor) Process(msg interface{}) {
    if setUp, ok := msg.(*bean.Setup); ok {
        if member := store.GetItSelfOnMember(); member != nil {
            member.ProcessSetUp(setUp)
        }
    }
}

type CommitProcessor struct {
}

func (p *CommitProcessor) Process(msg interface{}) {
    if commitment, ok := msg.(*bean.Commitment); ok {
        if leader := store.GetItSelfOnLeader(); leader != nil {
            leader.ProcessCommit(commitment)
        }
    }
}

type ChallengeProcessor struct {
}

func (p *ChallengeProcessor) Process(msg interface{}) {
    if challenge, ok := msg.(*bean.Challenge); ok {
        if member := store.GetItSelfOnMember(); member != nil {
            member.ProcessChallenge(challenge)
        }
    }
}

type ResponseProcessor struct {
}

func (p *ResponseProcessor) Process(msg interface{}) {
    if response, ok := msg.(*bean.Response); ok {
        if leader := store.GetItSelfOnLeader(); leader != nil {
            leader.ProcessResponse(response)
        }
    }
}
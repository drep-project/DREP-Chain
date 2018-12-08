package processor

import (
    "BlockChainTest/bean"
    "BlockChainTest/pool"
    "BlockChainTest/consensus/consmsg"
)

type SetUpProcessor struct {
}

func (p *SetUpProcessor) process(peer *bean.Peer, msg interface{}) {
    if setUp, ok := msg.(*bean.Setup); ok {
        //if member := store.GetItSelfOnMember(); member != nil {
        //    if !member.ProcessSetUp(setUp) {
        //        log.Println("SOS Help ", *setUp)
        //        store.SetRemainingSetup(setUp)
        //    }
        //} else {
        //    store.SetRemainingSetup(setUp)
        //}
        pool.Push(&consmsg.SetupMsg{Msg:setUp, Peer:peer})
    }
}

type CommitProcessor struct {}

func (p *CommitProcessor) process(peer *bean.Peer, msg interface{}) {
    if commitment, ok := msg.(*bean.Commitment); ok {
        //if leader := store.GetItSelfOnLeader(); leader != nil {
        //    leader.ProcessCommit(commitment)
        //}
        pool.Push(&consmsg.CommitmentMsg{Msg: commitment, Peer:peer})
    }
}

type ChallengeProcessor struct {}

func (p *ChallengeProcessor) process(peer *bean.Peer, msg interface{}) {
    if challenge, ok := msg.(*bean.Challenge); ok {
        //if member := store.GetItSelfOnMember(); member != nil {
        //    member.ProcessChallenge(challenge)
        //}
        pool.Push(&consmsg.ChallengeMsg{Msg:challenge, Peer:peer})
    }
}

type ResponseProcessor struct {}

func (p *ResponseProcessor) process(peer *bean.Peer, msg interface{}) {
    if response, ok := msg.(*bean.Response); ok {
        //if leader := store.GetItSelfOnLeader(); leader != nil {
        //    leader.ProcessResponse(response)
        //}
        pool.Push(&consmsg.ResponseMsg{Msg:response, Peer:peer})
    }
}
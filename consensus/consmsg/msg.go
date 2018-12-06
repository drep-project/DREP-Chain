package consmsg

import (
    "BlockChainTest/bean"
    "BlockChainTest/network"
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

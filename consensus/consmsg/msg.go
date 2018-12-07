package consmsg

import (
    "BlockChainTest/bean"
)

type SetupMsg struct {
    Msg *bean.Setup
    Peer *bean.Peer
}

type CommitmentMsg struct {
    Msg *bean.Commitment
    Peer *bean.Peer
}

type ChallengeMsg struct {
    Msg *bean.Challenge
    Peer *bean.Peer
}

type ResponseMsg struct {
    Msg *bean.Response
    Peer *bean.Peer
}

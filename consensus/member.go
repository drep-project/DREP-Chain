package consensus

import (
    "BlockChainTest/network"
    "BlockChainTest/bean"
    "BlockChainTest/mycrypto"
    "math/big"
    "BlockChainTest/log"
    "BlockChainTest/pool"
    "time"
    "BlockChainTest/consensus/consmsg"
)

type Member struct {
    leader *network.Peer
    prvKey *mycrypto.PrivateKey
    msg []byte

    k []byte
    r *big.Int

}

func NewMember(leader *network.Peer, prvKey *mycrypto.PrivateKey) *Member {
    m := &Member{}
    m.leader = leader
    m.prvKey = prvKey
    return m
}

func (m *Member) ProcessConsensus(f func(setup *bean.Setup)bool) []byte {
    log.Println("Member set up wait")
    if !m.waitForSetUp(f) {
        return nil
    }
    log.Println("Member is going to commit")
    m.commit()

    log.Println("Member challenge wait")
    if m.waitForChallenge() {
        log.Println("Member is going to response")
        m.response()
        return m.msg
    } else {
        return nil
    }
}

func (m *Member) waitForSetUp(f func(setup *bean.Setup)bool) bool {
    setUpMsg := pool.ObtainOne(func(msg interface{}) bool {
        if setup, ok := msg.(*consmsg.SetupMsg); ok {
            return m.leader.PubKey.Equal(setup.Peer.PubKey)
        } else {
            return false
        }
    }, 5 * time.Second)
    if setUpMsg == nil {
        return false
    }
    if setUp, ok := setUpMsg.(*consmsg.SetupMsg); ok {
        m.msg = setUp.Msg.Msg
        return f(setUp.Msg)
    } else {
        return false
    }
}

func (m *Member) commit()  {
    k, q, err := mycrypto.GetRandomKQ()
    if err != nil {
        return
    }
    m.k = k
    commitment := &bean.Commitment{Q: q}
    log.Println("Member commit ", *commitment)
    network.SendMessage([]*network.Peer{m.leader}, commitment)
}

func (m *Member) waitForChallenge() bool {
    challengeMsg := pool.ObtainOne(func(msg interface{}) bool {
        if challengeMsg, ok := msg.(*consmsg.ChallengeMsg); ok {
           return m.leader.PubKey.Equal(challengeMsg.Peer.PubKey)
        } else {
           return false
        }
    }, 5 * time.Second)
    if challengeMsg == nil {
        return false
    }
    if challenge, ok := challengeMsg.(*consmsg.ChallengeMsg); ok {
        log.Println("Member process challenge ", *challenge)
        r := mycrypto.ConcatHash256(challenge.Msg.SigmaQ.Bytes(), challenge.Msg.SigmaPubKey.Bytes(), m.msg)
        r0 := new(big.Int).SetBytes(challenge.Msg.R)
        rInt := new(big.Int).SetBytes(r)
        curve := mycrypto.GetCurve()
        rInt.Mod(rInt, curve.N)
        m.r = rInt
        return r0.Cmp(m.r) == 0
    } else {
        return false
    }
}

func (m *Member) response() {
    curve := mycrypto.GetCurve()
    k := new(big.Int).SetBytes(m.k)
    prvInt := new(big.Int).SetBytes(m.prvKey.Prv)
    s := new(big.Int).Mul(m.r, prvInt)
    s.Sub(k, s)
    s.Mod(s, curve.N)
    response := &bean.Response{S: s.Bytes()}
    log.Println("Member response ", *response)
    network.SendMessage([]*network.Peer{m.leader}, response)
}
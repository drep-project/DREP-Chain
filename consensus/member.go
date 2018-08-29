package consensus

import (
    "BlockChainTest/network"
    "sync"
    "BlockChainTest/bean"
    "BlockChainTest/crypto"
    "math/big"
    "BlockChainTest/hash"
    "fmt"
)

type Member struct {
    leader *network.Peer
    state int
    prvKey *bean.PrivateKey
    pubKey *bean.Point
    msg []byte

    k []byte
    r *big.Int

    setUpWg sync.WaitGroup
    challengeWg sync.WaitGroup

}

func NewMember(leader *network.Peer, prvKey *bean.PrivateKey, pubKey *bean.Point) *Member {
    m := &Member{}
    m.state = waiting
    m.leader = leader
    m.prvKey = prvKey
    m.pubKey = pubKey
    return m
}
func (m *Member) ProcessConsensus() {
    m.setUpWg = sync.WaitGroup{}
    m.setUpWg.Add(1)
    fmt.Println("Member set up wait")
    m.setUpWg.Wait()
    fmt.Println("Member is going to commit")
    m.commit()

    m.challengeWg = sync.WaitGroup{}
    m.challengeWg.Add(1)
    fmt.Println("Member challenge wait")
    m.challengeWg.Wait()
    fmt.Println("Member is going to response")
    m.response()
}

func (m *Member) ProcessSetUp(setupMsg *bean.Setup) {
    //if !store.CheckRole(node.MINER) {
    //    return
    //}
    if !m.leader.PubKey.Equal(setupMsg.PubKey) {
        return
    }
    if m.state != waiting {
        return
    }
    fmt.Println("Member process setup ", *setupMsg)
    m.msg = setupMsg.Msg
    m.setUpWg.Done()
}

func (m *Member) commit()  {
    k, q, err := crypto.GetRandomKQ()
    if err != nil {
        return
    }
    pubKey := m.pubKey
    m.k = k
    commitment := &bean.Commitment{PubKey: pubKey, Q: q}
    fmt.Println("Member commit ", *commitment)
    network.SendMessage([]*network.Peer{m.leader}, commitment)
}

func (m *Member) ProcessChallenge(challenge *bean.Challenge) {
    fmt.Println("Member process challenge ", *challenge)
    r := hash.ConcatHash256(challenge.SigmaQ.Bytes(), challenge.SigmaPubKey.Bytes(), m.msg)
    r0 := new(big.Int).SetBytes(challenge.R)
    rInt := new(big.Int).SetBytes(r)
    curve := crypto.GetCurve()
    rInt.Mod(rInt, curve.N)
    m.r = rInt
    if r0.Cmp(m.r) != 0 {
        m.challengeWg.Done()
        return// errors.New("wrong hash value")
    }
    m.challengeWg.Done()
}

func (m *Member) response()  {
    curve := crypto.GetCurve()
    prvKey := m.prvKey
    k := new(big.Int).SetBytes(m.k)
    prvInt := new(big.Int).SetBytes(prvKey.Prv)
    s := new(big.Int).Mul(m.r, prvInt)
    s.Sub(k, s)
    s.Mod(s, curve.N)
    response := &bean.Response{PubKey: prvKey.PubKey, S: s.Bytes()}
    peers := make([]*network.Peer, 1)
    peers[0] = m.leader
    fmt.Println("Member response ", *response)
    network.SendMessage(peers, response)
}
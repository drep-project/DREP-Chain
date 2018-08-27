package consensus

import (
    "BlockChainTest/store"
    "BlockChainTest/node"
    "BlockChainTest/network"
    "sync"
    "BlockChainTest/bean"
    "BlockChainTest/crypto"
    "math/big"
)

type Member struct {
    leader *node.Miner
    state int
    prvKey *bean.PrivateKey
    msg []byte

    k []byte
    r *big.Int

    setUpWg sync.WaitGroup
    challengeWg sync.WaitGroup

}

func NewMember() *Member {
    m := &Member{}
    m.state = waiting
    m.leader = store.GetLeader()
    m.prvKey = store.GetPriKey()
    return m
}
func (m *Member) processConsensus() {
    m.setUpWg = sync.WaitGroup{}
    m.setUpWg.Add(1)
    m.setUpWg.Wait()

    m.commit()

    m.challengeWg = sync.WaitGroup{}
    m.challengeWg.Add(1)
    m.challengeWg.Wait()

    m.response()
}

func (m *Member) ProcessSetUp(setupMsg *bean.Setup) {
    if !store.CheckRole(node.MINER) {
        return
    }
    if !store.GetLeader().PubKey.Equal(setupMsg.PubKey) {
        return
    }
    if m.state != waiting {
        return
    }
    m.msg = setupMsg.Msg
    m.setUpWg.Done()
}

func (m *Member) commit()  {
    k, q, err := crypto.GetRandomKQ()
    if err != nil {
        return
    }
    pubKey := store.GetPubKey()
    m.k = k
    commitment := &bean.Commitment{PubKey: pubKey, Q: q}
    network.SendMessage([]*network.Peer{m.leader.Peer}, commitment)
}

func (m *Member) ProcessChallenge(challenge *bean.Challenge) {
    r := crypto.ConcatHash256(challenge.SigmaQ.Bytes(), challenge.SigmaPubKey.Bytes(), m.msg)
    r0 := new(big.Int).SetBytes(challenge.R)
    m.r = new(big.Int).SetBytes(r)
    if r0.Cmp(m.r) != 0 {
        return// errors.New("wrong hash value")
    }
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
    peers[0] = m.leader.Peer
    network.SendMessage(peers, response)
}
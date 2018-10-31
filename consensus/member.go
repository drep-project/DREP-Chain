package consensus

import (
    "BlockChainTest/network"
    "sync"
    "BlockChainTest/bean"
    "BlockChainTest/mycrypto"
    "math/big"
    "BlockChainTest/log"
    "BlockChainTest/pool"
    "time"
)

type Member struct {
    leader *network.Peer
    state int
    prvKey *mycrypto.PrivateKey
    pubKey *mycrypto.Point
    msg []byte

    k []byte
    r *big.Int

    challengeWg sync.WaitGroup

}

func NewMember(leader *network.Peer, prvKey *mycrypto.PrivateKey) *Member {
    m := &Member{}
    m.state = waiting
    m.leader = leader
    m.prvKey = prvKey
    m.pubKey = prvKey.PubKey

    m.challengeWg = sync.WaitGroup{}
    m.challengeWg.Add(1)
    return m
}
func (m *Member) ProcessConsensus() []byte {
    log.Println("Member set up wait")
    m.waitForSetUp()
    log.Println("Member is going to commit")
    m.commit()

    log.Println("Member challenge wait")
    m.challengeWg.Wait()
    log.Println("Member is going to response")
    m.response()
    return m.msg
}

func (m *Member) waitForSetUp() bool {
    setUpMsg := pool.ObtainOne(func(msg interface{}) bool {
        if setup, ok := msg.(*bean.Setup); ok {
            return m.leader.PubKey.Equal(setup.PubKey)
        } else {
            return false
        }
    }, 5 * time.Second)
    if setUpMsg == nil {
        return false
    }
    if setUp, ok := setUpMsg.(*bean.Setup); ok {
        m.msg = setUp.Msg
        return true
    } else {
        return false
    }
}

func (m *Member) commit()  {
    k, q, err := mycrypto.GetRandomKQ()
    if err != nil {
        return
    }
    pubKey := m.pubKey
    m.k = k
    commitment := &bean.Commitment{PubKey: pubKey, Q: q}
    log.Println("Member commit ", *commitment)
    network.SendMessage([]*network.Peer{m.leader}, commitment)
}

func (m *Member) ProcessChallenge(challenge *bean.Challenge) {
    log.Println("Member process challenge ", *challenge)
    r := mycrypto.ConcatHash256(challenge.SigmaQ.Bytes(), challenge.SigmaPubKey.Bytes(), m.msg)
    r0 := new(big.Int).SetBytes(challenge.R)
    rInt := new(big.Int).SetBytes(r)
    curve := mycrypto.GetCurve()
    rInt.Mod(rInt, curve.N)
    m.r = rInt
    if r0.Cmp(m.r) != 0 {
        m.challengeWg.Done()
        return// errors.New("wrong hash value")
    }
    m.challengeWg.Done()
}

func (m *Member) response()  {
    curve := mycrypto.GetCurve()
    prvKey := m.prvKey
    k := new(big.Int).SetBytes(m.k)
    prvInt := new(big.Int).SetBytes(prvKey.Prv)
    s := new(big.Int).Mul(m.r, prvInt)
    s.Sub(k, s)
    s.Mod(s, curve.N)
    response := &bean.Response{PubKey: prvKey.PubKey, S: s.Bytes()}
    log.Println("Member response ", *response)
    network.SendMessage([]*network.Peer{m.leader}, response)
}
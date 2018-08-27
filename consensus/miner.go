package processor

import (
    "BlockChainTest/common"
    "BlockChainTest/store"
    "BlockChainTest/node"
    "fmt"
    "bytes"
    "BlockChainTest/network"
    "sync"
    "BlockChainTest/bean"
    "BlockChainTest/crypto"
    "math/big"
    "errors"
)

type Miner struct {
    Leader *node.Miner
    state int
    PrvKey *bean.PrivateKey
    Msg []byte

    K []byte
    sigmaS *big.Int
    responseWg sync.WaitGroup
    responseBitmap map[*bean.Point]bool

    sigs map[common.Address][]byte

    setUpWg sync.WaitGroup
}

func NewMiner() *Miner {
    m := &Miner{}
    m.state = waiting
    return m
}
func (m *Miner) processConsensus() {
    m.commitWg = sync.WaitGroup{}
    l.commitWg.Add(len(l.peers))
    l.state = setUp
    l.setUp(msg, store.GetPubKey())
    l.commitWg.Wait()

    l.responseWg = sync.WaitGroup{}
    l.responseWg.Add(len(l.commitBitmap))
    l.state = challenge
    l.challenge(msg)
    l.responseWg.Wait()

    return &bean.Signature{R: l.r, S: l.sigmaS.Bytes()}
}

func (m *Miner) processSetUp(setupMsg *bean.Setup) {
    if !store.CheckRole(node.MINER) {
        return
    }
    if !store.GetLeader().PubKey.Equal(setupMsg.PubKey) {
        return
    }
    if m.state != waiting {
        return
    }
    k, q, err := crypto.GetRandomKQ()
    if err != nil {
        return
    }
    pubKey := store.GetPubKey()
    m.K = k
    commitment := &bean.Commitment{PubKey: pubKey, Q: q}
    network.SendMessage([]*network.Peer{m.Leader.Peer}, commitment)
}

func (m *Miner) processChallenge(challenge *bean.Challenge) error {
    curve := crypto.GetCurve()
    prvKey := m.PrvKey
    r := crypto.ConcatHash256(challenge.SigmaQ.Bytes(), challenge.SigmaPubKey.Bytes(), m.Msg)
    r0 := new(big.Int).SetBytes(challenge.R)
    r1 := new(big.Int).SetBytes(r)
    if r0.Cmp(r1) != 0 {
        return errors.New("wrong hash value")
    }
    k := new(big.Int).SetBytes(m.K)
    prvInt := new(big.Int).SetBytes(prvKey.Prv)
    s := new(big.Int).Mul(r1, prvInt)
    s.Sub(k, s)
    s.Mod(s, curve.N)
    response := &bean.Response{PubKey: prvKey.PubKey, S: s.Bytes()}
    peers := make([]*network.Peer, 1)
    peers[0] = m.Leader.Peer
    network.SendMessage(peers, response)
    return nil
}

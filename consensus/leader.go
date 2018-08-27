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

const (
    waiting              = 0
    setUp               = 1
    challenge            = 2
)
type Leader struct {
    miners []*node.Miner
    peers []*network.Peer
    state int
    LeaderPeer *network.Peer

    commitWg sync.WaitGroup
    commitBitmap map[string] bool
    sigmaPubKey *bean.Point
    sigmaQ *bean.Point
    r []byte

    sigmaS *big.Int
    responseWg sync.WaitGroup
    responseBitmap map[string] bool

    sigs map[bean.Address][]byte

}

func NewLeader() *Leader {
    l := &Leader{}
    miners := store.GetMiners()
    l.miners = make([]*node.Miner, len(miners) - 1)
    last := 0
    pubKey := store.GetPubKey()
    for _, miner := range miners {
        if !miner.PubKey.Equal(pubKey) {
            l.miners[last] = miner
            last++
        }
    }
    l.state = waiting
    return l
}

func (l *Leader) processConsensus(msg []byte) *bean.Signature {
    l.commitWg = sync.WaitGroup{}
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

func (l *Leader) setUp(msg []byte, pubKey *bean.Point) {
    setup := &bean.Setup{Msg: msg, PubKey: pubKey}
    network.SendMessage(l.peers, setup)
}

func (l *Leader) getR(msg []byte) []byte {
    curve := crypto.GetCurve()
    r := crypto.ConcatHash256(l.sigmaQ.Bytes(), l.sigmaPubKey.Bytes(), msg)
    rInt := new(big.Int).SetBytes(r)
    rInt.Mod(rInt, curve.N)
    return rInt.Bytes()
}

func (l *Leader) challenge(msg []byte)  {
    l.r = l.getR(msg)
    challenge := &bean.Challenge{SigmaPubKey: l.sigmaPubKey, SigmaQ: l.sigmaQ, R: l.r}
    network.SendMessage(l.peers, challenge)
}

func (l *Leader) processCommit(commit *bean.Commitment) {
    if l.state != setUp {
        return
    }
    if !store.CheckRole(node.LEADER) {
        return
    }
    addr := commit.PubKey.Addr()
    if l.commitBitmap[addr] {
       return
    }
    l.commitBitmap[addr] = true
    l.commitWg.Done()
    curve := crypto.GetCurve()
    l.sigmaPubKey = curve.Add(l.sigmaPubKey, commit.PubKey)
    l.sigmaQ = curve.Add(l.sigmaQ, commit.Q)
}

func (l *Leader) processResponse(response *bean.Response) {
    if l.state != challenge {
        return
    }
    if !store.CheckRole(node.LEADER) {
        return
    }
    addr := response.PubKey.Addr()
    if l.responseBitmap[addr] {
       return
    }
    l.responseBitmap[addr] = true
    l.responseWg.Done()
    s := new(big.Int).SetBytes(response.S)
    l.sigmaS = l.sigmaS.Add(l.sigmaS, s)
}
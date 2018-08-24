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

}

func NewMiner() *Leader {
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
func (m *Miner) processConsensus(msg interface{},
                                 sigFunc func([]byte, *bean.Point) *bean.Signature)  {
    if !store.CheckState(node.MINER, common.WAITING) {
        return
    }
    if setup1Msg, ok := msg.(common.SetUp1Message); ok {
        fmt.Println(setup1Msg)
        // TODO Check sig
        if !bytes.Equal(store.GetLeader().PubKey, setup1Msg.PubKey) {
            return
        }
        if setup1Msg.BlockHeight != store.GetBlockHeight() + 1 {
            return
        }
        store.SetBlock(setup1Msg.Block)
        store.MoveToState(common.MSG_BLOCK1_RESPONSE)
        // TODO clear block1CommitProcessor and Start countdown
        // TODO Get Qi
        //q := crypto.GetQ()
        peer := store.GetLeader().Peer
        // TODO Send Qi to the leader
        // TODO Generate the block
        //network.SendMessage(peer, block1CommitMsg{q, pubKey})
    }
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
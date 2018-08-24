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
    leader *node.Miner
    state int

    k []byte
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
    m.k = k
    commitment := &bean.Commitment{PubKey: pubKey, Q: q}
    network.SendMessage([]*network.Peer{m.leader.Peer}, commitment)
}

func (m *Miner) processChallenge(msg interface{}) {
    if !store.CheckState(node.MINER, common.MSG_BLOCK1_RESPONSE) {
        return
    }
    if block1ResponseMsg, ok := msg.(common.Block1ResponseMessage); ok {
        fmt.Println(block1ResponseMsg)
        miner := store.GetMiner(block1ResponseMsg.PubKey)
        if miner == nil {
            return
        }
        // TODO calculate s
        // TODO send s to leader
    }
}
package node

import (
    "BlockChainTest/store"
    "BlockChainTest/bean"
    "BlockChainTest/consensus"
    "github.com/golang/protobuf/proto"
    "sync"
    "BlockChainTest/log"
)

const (
    //LEADER     = 0
    //MEMBER1      = 1
    //MEMBER2 = 2
    //NON_MINER  = 3
    LEADER    = 0
    MEMBER    = 1
    OTHER     = 2
)

type Node struct {
    role int
    address *bean.Address
    prvKey *bean.PrivateKey
    wg *sync.WaitGroup
}

func NewNode(role int, prvKey *bean.PrivateKey) *Node {
    address := prvKey.PubKey.Addr()
    return &Node{role: role, address: &address, prvKey: prvKey}
}

func (n *Node) isLeader() bool {
    if leader := store.GetLeader(); leader != nil {
        return *n.address == leader.Address
    } else {
        return false
    }
}

func (n *Node) Start() {
    for {
        log.Println("node start")
        switch store.GetRole() {
        case LEADER:
            n.runAsLeader()
        case MEMBER:
            n.runAsMember()
        case OTHER:
            n.runAsOther()
        }
        log.Println("node stop")
    }
}

func (n *Node) runAsLeader() {
    leader1 := consensus.NewLeader(n.prvKey.PubKey, store.GetMiners())
    block := store.GetBlock()
    log.Println("node leader is preparing process consensus for round 1")
    if msg, err := proto.Marshal(block); err ==nil {
        log.Println("node leader is going to process consensus for round 1")
        sig, bitmap := leader1.ProcessConsensus(msg)
        multiSig := &bean.MultiSignature{Sig: sig, Bitmap: bitmap}
        log.Println("node leader is preparing process consensus for round 2")
        if msg, err := proto.Marshal(multiSig); err == nil {
            leader2 := consensus.NewLeader(n.prvKey.PubKey, store.GetMiners())
            log.Println("node leader is going to process consensus for round 2")
            leader2.ProcessConsensus(msg)
            log.Println("node leader finishes process consensus for round 2")
        }
    }
}

func (n *Node) runAsMember() {
    member1 := consensus.NewMember(store.GetLeader(), store.GetPrvKey())
    log.Println("node member is going to process consensus for round 1")
    bytes := member1.ProcessConsensus()
    log.Println("node member finishes consensus for round 1")
    block := &bean.Block{}
    n.wg = &sync.WaitGroup{}
    n.wg.Add(1)
    if proto.Unmarshal(bytes, block) != nil {
        member2 := consensus.NewMember(store.GetLeader(), store.GetPrvKey())
        log.Println("node member is going to process consensus for round 2")
        member2.ProcessConsensus()
        log.Println("node member finishes consensus for round 2")
    }
    log.Println("node member is going to wait")
    n.wg.Wait()
    log.Println("node member finishes wait")
}

func (n *Node) runAsOther() {

}

func (n *Node) processBlock(block *bean.Block) {
    log.Println("Receive ", *block)
    n.wg.Done()
}
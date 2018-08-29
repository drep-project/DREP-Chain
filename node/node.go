package node

import (
    "BlockChainTest/store"
    "BlockChainTest/bean"
    "BlockChainTest/consensus"
    "github.com/golang/protobuf/proto"
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
        switch store.GetRole() {
        case LEADER:
            n.runAsLeader()
        case MEMBER:
            n.runAsMember()
        case OTHER:
            n.runAsOther()
        }
    }
}

func (n *Node) runAsLeader() {
    leader := consensus.NewLeader(n.prvKey.PubKey, store.GetMiners())
    block := store.GetBlock()
    if msg, err := proto.Marshal(block); err ==nil {
        sig, bitmap := leader.ProcessConsensus(msg)
        multiSig := &bean.MultiSignature{Sig: sig, Bitmap: bitmap}
        if msg, err := proto.Marshal(multiSig); err == nil {
            leader.ProcessConsensus(msg)
        }
    }
}

func (n *Node) runAsMember() {
    n.address
}

func (n *Node) runAsOther() {

}

func (n *Node) processBlock(block *bean.Block) {

}
package node

import (
    "BlockChainTest/store"
    "BlockChainTest/bean"
    "BlockChainTest/consensus"
    "github.com/golang/protobuf/proto"
    "sync"
    "BlockChainTest/log"
    "BlockChainTest/network"
    "BlockChainTest/crypto"
    "time"
    "BlockChainTest/role"
)

var (
    once sync.Once
    node *Node
)

type Node struct {
    address *bean.Address
    prvKey *crypto.PrivateKey
    wg *sync.WaitGroup
}

func newNode(prvKey *crypto.PrivateKey) *Node {
    address := bean.Addr(prvKey.PubKey)
    return &Node{address: &address, prvKey: prvKey}
}

func GetNode() *Node {
    once.Do(func() {
        node = newNode(store.GetPrvKey())
    })
    return node
}

func (n *Node) isLeader() bool {
    if leader := store.GetLeader(); leader != nil {
        return *n.address == leader.Address
    } else {
        return false
    }
}

func (n *Node) Start() {
    go func() {
        for {
            time.Sleep(5 * time.Second)
            log.Println("node start")
            store.ChangeRole()
            switch store.GetRole() {
            case bean.LEADER:
                n.runAsLeader()
            case bean.MEMBER:
                n.runAsMember()
            case bean.NEWCOMER:
                n.runAsNewComer()
            case bean.OTHER:
                n.runAsOther()
            }
            log.Println("node stop")
            log.Println("Current height ", store.GetCurrentBlockHeight())
        }
    }()
}

//TODO : simulate the newcomer join in
// func (n *Node) Start()  {
//     store.NewcomerRole()
//     switch store.GetRole() {
//     case bean.NEWCOMER:
//         n.runAsNewComer()
//     case bean.OTHER:
//         n.runAsOther()
//     }
// }

func (n *Node) runAsLeader() {
    leader1 := consensus.NewLeader(n.prvKey.PubKey, store.GetMiners())
    store.SetLeader(leader1)
    block := store.GenerateBlock()
    log.Println("node leader is preparing process consensus for round 1")
    if msg, err := proto.Marshal(block); err ==nil {
        log.Println("node leader is going to process consensus for round 1")
        sig, bitmap := leader1.ProcessConsensus(msg)
        multiSig := &bean.MultiSignature{Sig: sig, Bitmap: bitmap}
        log.Println("node leader is preparing process consensus for round 2")
        if msg, err := proto.Marshal(multiSig); err == nil {
            leader2 := consensus.NewLeader(n.prvKey.PubKey, store.GetMiners())
            store.SetLeader(leader2)
            log.Println("node leader is going to process consensus for round 2")
            leader2.ProcessConsensus(msg)
            log.Println("node leader finishes process consensus for round 2")
            log.Println("node leader is going to send block")
            n.sendBlock(block)
            n.ProcessBlock(block, false)
            log.Println("node leader finishes sending block")
        }
    }
}

func (n *Node) sendBlock(block *bean.Block) {
    peers := store.GetPeers()
    network.SendMessage(peers, block)
}

func (n *Node) runAsMember() {
    member1 := consensus.NewMember(store.GetLeader(), store.GetPrvKey())
    store.SetMember(member1)
    log.Println("node member is going to process consensus for round 1")
    setUp := store.GetRemainingSetup()
    var bytes []byte
    if setUp != nil {
        bytes = member1.ProcessConsensus(setUp, func() {
            store.SetRemainingSetup(nil)
        })
    } else {
        bytes = member1.ProcessConsensus(nil, nil)
    }
    log.Println("node member finishes consensus for round 1")
    block := &bean.Block{}
    n.wg = &sync.WaitGroup{}
    n.wg.Add(1)
    if proto.Unmarshal(bytes, block) == nil {
        member2 := consensus.NewMember(store.GetLeader(), store.GetPrvKey())
        store.SetMember(member2)
        log.Println("node member is going to process consensus for round 2")
        member2.ProcessConsensus(nil, nil)
        log.Println("node member finishes consensus for round 2")
    }
    log.Println("node member is going to wait")
    n.wg.Wait()
    log.Println("node member finishes wait")
}

func (n *Node) runAsOther() {
    if n.prvKey != nil {
       n.runAsUser()
       return
    }
}

func (n *Node) runAsUser()  {
    peer := store.GetPeers()[0]
    user := role.NewUser(peer.PubKey, store.GetPeers())
    store.SetUser(user)
    log.Println("already ran as user!")
}

func (n *Node) runAsNewComer() {
    // TODO: a hard coding server
    curve := crypto.GetCurve()
    k := []byte{0x22, 0x11}
    pub := curve.ScalarBaseMultiply(k)
    ip := network.IP("192.168.3.13")
    port := network.Port(55555)
    peer := &network.Peer{IP: ip, Port: port, PubKey: pub}

    newcomer := role.NewJoiner(peer)
    store.SetNewComer(newcomer)
    log.Println("newcomer is going to process")
    newcomer.ProcessJoin()
    log.Println("welcome the newcomer!")
}

func (n *Node) ProcessBlock(block *bean.Block, del bool) {
    log.Println("node receive block", *block)
    store.ExecuteTransactions(block, del)
    if del {
        n.wg.Done()
    }
}
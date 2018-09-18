package node

import (
    "BlockChainTest/store"
    "BlockChainTest/bean"
    "BlockChainTest/consensus"
    "github.com/golang/protobuf/proto"
    "sync"
    "BlockChainTest/log"
    "BlockChainTest/network"
    "BlockChainTest/mycrypto"
    "time"
    "fmt"
    "container/list"
)

var (
    once sync.Once
    node *Node
)

type Node struct {
    address *bean.Address
    prvKey *mycrypto.PrivateKey
    wg *sync.WaitGroup
    prep  bool
    prepLock sync.Mutex
    prepCond *sync.Cond
}

func newNode(prvKey *mycrypto.PrivateKey) *Node {
    address := bean.Addr(prvKey.PubKey)
    n := &Node{address: &address, prvKey: prvKey, prep:false}
    n.prepCond = sync.NewCond(&n.prepLock)
    return n
}

func GetNode() *Node {
    once.Do(func() {
        node = newNode(store.GetPrvKey())
    })
    return node
}

func (n *Node) Start() {
    if store.IsAdmin {

    }
    go func() {
        for {
            time.Sleep(5 * time.Second)
            log.Println("node start")
            isM, isL := store.MoveToNextMiner();
            if isL {
                n.runAsLeader()
            } else {
                n.wg = &sync.WaitGroup{}
                n.wg.Add(1)
                n.prepLock.Lock()
                n.prep = true
                n.prepCond.Broadcast()
                n.prepLock.Unlock()
                if isM {
                    n.runAsMember()
                }
                n.wg.Wait() // If not, next will be nil member
            }
            log.Println("node stop")
            log.Println("Current height ", store.GetCurrentBlockHeight())
        }
    }()
}

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
            n.ProcessBlock(block, false) // process before sending
            n.sendBlock(block)
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
    //n.wg = &sync.WaitGroup{}
    //n.wg.Add(1)
    if proto.Unmarshal(bytes, block) == nil {
        member2 := consensus.NewMember(store.GetLeader(), store.GetPrvKey())
        store.SetMember(member2)
        log.Println("node member is going to process consensus for round 2")
        member2.ProcessConsensus(nil, nil)
        log.Println("node member finishes consensus for round 2")
    }
    //log.Println("node member is going to wait")
    //n.wg.Wait()
    //log.Println("node member finishes wait")
}

func (n *Node) ProcessBlock(block *bean.Block, del bool) {
    if del {
        n.prepLock.Lock()
        for !n.prep {
            n.prepCond.Wait()
        }
        n.prepLock.Unlock()
    }
    log.Println("node receive block", *block)
    fmt.Println("Process block leader = ", bean.Addr(block.Header.LeaderPubKey))
    store.ExecuteTransactions(block, del)
    if del {
        n.wg.Done()
    }
}

func (n *Node) discover() {
    msg := &bean.PeerInfo{Pk: n.prvKey.PubKey, Ip:"192.168.3.113", Port: 55555}
    peers := []*network.Peer{store.Admin}
    network.SendMessage(peers, msg)
}

func (n *Node) ProcessNewPeer(newcomer *bean.PeerInfo) {
    log.Println("user starting process a newcomer")
    peers := store.GetPeers()
    newPeer := &network.Peer{
        IP:network.IP(newcomer.Ip),
        Port: network.Port(newcomer.Port),
        PubKey: newcomer.Pk}
    store.AddPeer(newPeer)
    list := make([]*bean.PeerInfo, 0)
    for _, p := range peers {
        t := &bean.PeerInfo{Pk: p.PubKey, Ip:string(p.IP), Port:int32(p.Port)}
        list = append(list, t)
    }
    peerList := &bean.PeerInfoList{List:list}
    network.SendMessage([]*network.Peer{newPeer}, peerList)
    network.SendMessage(store.GetPeers(), newcomer)
}

func (n *Node) ProcessPeerList(list *bean.PeerInfoList) {
    for _, t := range list.List {
        store.AddPeer(&network.Peer{IP:network.IP(t.Ip), Port:network.Port(t.Port), PubKey:t.Pk})
    }
}
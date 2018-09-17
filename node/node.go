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
)

var (
    once sync.Once
    node *Node
)

type Node struct {
    address *bean.Address
    prvKey *mycrypto.PrivateKey
    wg *sync.WaitGroup
}

func newNode(prvKey *mycrypto.PrivateKey) *Node {
    address := bean.Addr(prvKey.PubKey)
    return &Node{address: &address, prvKey: prvKey}
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
    if store.GetRole() == bean.MINER {
        go func() {
            for {
                time.Sleep(5 * time.Second)
                log.Println("node start")
                if store.MoveToNextMiner() {
                    n.runAsLeader()
                } else {
                    n.runAsMember()
                }
                log.Println("node stop")
                log.Println("Current height ", store.GetCurrentBlockHeight())
            }
        }()
    }
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

func (n *Node) ProcessBlock(block *bean.Block, del bool) {
    log.Println("node receive block", *block)
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

func (n *Node) ProcessMinerInfo(minerInfo *bean.MinerInfo) {
    peerInfo := minerInfo.Peer
    peer := &network.Peer{IP: network.IP(peerInfo.Ip), Port: network.Port(peerInfo.Port), PubKey:peerInfo.Pk}
    store.AddMiner(peer)
}
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
    "fmt"
    "container/list"
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
    if store.IsAdmin {

    }
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
            case bean.OTHER:
                n.runAsOther()
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

}

func (n *Node) ProcessBlock(block *bean.Block, del bool) {
    log.Println("node receive block", *block)
    store.ExecuteTransactions(block, del)
    if del {
        n.wg.Done()
    }
}

func (n *Node) discover() {
    msg := &bean.Newcomer{Pk: n.prvKey.PubKey, Ip:"192.168.3.113", Port: 55555}
    peers := []*network.Peer{store.Admin}
    network.SendMessage(peers, msg)
}

func (n *Node) ProcessNewComer(newcomer *bean.Newcomer)  {
    log.Println("user starting process a newcomer")
    pubKey := newcomer.Pk
    address := bean.Addr(pubKey)
    newPeer := &network.Peer{}
    newPeer.IP = network.IP(newcomer.Ip)
    newPeer.Port = network.Port(newcomer.Port)

    newPeer.PubKey = pubKey
    newPeer.Address = address
    peerStore := network.GetStore()
    peerStore.AddPeer(newPeer)

    list := make([]*bean.Newcomer, 0)
    peers := make([]*network.Peer, 0)
    for _, value := range peerStore.Store {
        msg := &bean.Newcomer{}
        msg.Pk = value.PubKey
        list = append(list, msg)

        peer := &network.Peer{}
        peer.PubKey = value.PubKey
        peer.Address = address
        peer.IP = value.IP
        peer.Port = value.Port
        peers = append(peers, peer)
    }

    listOfPeer := &bean.ListOfPeer{}
    listOfPeer.List = list

    // return the list to newcomer
    log.Println("send a list of peer to the newcomer")

    //newcomers := []*network.Peer{newPeer}
    //network.SendMessage(newcomers, listOfPeer)

    task := network.Task{newPeer,listOfPeer}
    task.sendMessageCore()

    // broadcast the new comer msg
    //network.SendMessage(n.peers, newcomer)
}

func (n *Node) ProcessPeers(list *bean.ListOfPeer) {
    log.Println("welcome newcomer! it's done.")
    peerStore := network.GetStore()
    fmt.Println("the peerStore before: ", peerStore.Store)
    // store the peers in the local memory.
    for _, item := range list.List {
        pubKey := item.Pk
        address := bean.Addr(pubKey)
        peer := &network.Peer{}
        peer.PubKey = pubKey
        peer.Address = address
        peerStore.AddPeer(peer)
        peerStore.Store[address] = peer
    }
    fmt.Println("the peerStore after: ", peerStore.Store)
    log.Println("newcomer has refreshed the peerStore")
    n.state = done
    n.wg.Done()
    log.Println("add newcomer done")
}
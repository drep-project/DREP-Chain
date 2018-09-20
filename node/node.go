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
    "fmt"
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
    discoverWg *sync.WaitGroup
    fetchLock sync.Mutex
    fetchCond *sync.Cond
    curMaxHeight int64
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
    if store.IsStart {

    } else {
        n.discover()
        n.fetchBlocks()
    }
    go func() {
        for {
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
            time.Sleep(5 * time.Second)
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
    fmt.Println("discovering 1")
    msg := &bean.PeerInfo{Pk: n.prvKey.PubKey, Ip:"192.168.3.113", Port: 55555}
    peers := []*network.Peer{store.Admin}
    fmt.Println("discovering 2")
    n.discoverWg = &sync.WaitGroup{}
    n.discoverWg.Add(1)
    network.SendMessage(peers, msg)
    fmt.Println("discovering 3")
    n.discoverWg.Wait()
    fmt.Println("discovering 4")
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
    fmt.Println("discovering 5")
    for _, t := range list.List {
        store.AddPeer(&network.Peer{IP:network.IP(t.Ip), Port:network.Port(t.Port), PubKey:t.Pk})
    }
    fmt.Println("discovering 6")
    n.discoverWg.Done()
    fmt.Println("discovering 7")
}

func (n *Node) fetchBlocks() {
    peers := store.GetPeers()
    if len(peers) == 0 {
        log.Errorf("Fuck")
        return
    }
    n.fetchCond = sync.NewCond(&n.fetchLock)
    n.fetchLock.Lock()
    defer n.fetchLock.Unlock()
    n.curMaxHeight = 2<<60
    req := &bean.BlockReq{Height:store.GetCurrentBlockHeight()}

    network.SendMessage([]*network.Peer{peers[0]}, req)
    fmt.Println("fetching 1")
    for n.curMaxHeight != store.GetCurrentBlockHeight() {
        fmt.Println("fetching 2: ", n.curMaxHeight, store.GetCurrentBlockHeight())
        n.fetchCond.Wait()
        fmt.Println("fetching 3")
    }
}

func (n *Node) ProcessBlockResp(resp *bean.BlockResp) {
    fmt.Println("fetching 4")
    for _, b := range resp.Blocks {
        n.ProcessBlock(b, false)
        // TODO cannot receive tran
    }
    fmt.Println("fetching 5")
    n.fetchLock.Lock()
    defer n.fetchLock.Unlock()
    fmt.Println("fetching 6 ", resp.Height)
    n.curMaxHeight = resp.Height
    n.fetchCond.Broadcast()
}

func (n *Node) ProcessBlockReq(req *bean.BlockReq) {
    from := req.Height + 1
    size := int64(2)
    peers := []*network.Peer{store.GetPeer(req.Pk)}
    fmt.Println("ProcessBlockReq")
    for i := from; i <= store.GetCurrentBlockHeight(); {
        fmt.Println("ProcessBlockReq 1 ", i)
        bs := store.GetBlocks(i, size)
        resp := &bean.BlockResp{Height:store.GetCurrentBlockHeight(), Blocks:bs}
        network.SendMessage(peers, resp)
        i += int64(len(bs))
        fmt.Println("ProcessBlockReq 2 ", i)
    }
}


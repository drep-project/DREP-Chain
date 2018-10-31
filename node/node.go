package node

import (
    "BlockChainTest/store"
    "BlockChainTest/bean"
    "BlockChainTest/consensus"
    "sync"
    "BlockChainTest/log"
    "BlockChainTest/network"
    "BlockChainTest/mycrypto"
    "fmt"
    "time"
    "math/big"
    "BlockChainTest/util/concurrent"
    "BlockChainTest/util"
    "BlockChainTest/database"
    "encoding/json"
)

var (
    once sync.Once
    node *Node
)

type Node struct {
    prvKey *mycrypto.PrivateKey
    wg concurrent.CountDownLatch
    discoverWg *sync.WaitGroup
    fetchLock sync.Mutex
    fetchCond *sync.Cond
    curMaxHeight int64
    prep  bool
    prepLock sync.Mutex
    prepCond *sync.Cond
    discovering bool
    pingLatches map[bean.Address]concurrent.CountDownLatch
}

func newNode(prvKey *mycrypto.PrivateKey) *Node {
    n := &Node{prvKey: prvKey, prep:false}
    n.prepCond = sync.NewCond(&n.prepLock)
    n.pingLatches = make(map[bean.Address]concurrent.CountDownLatch)
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
        n.discovering = false
    } else {
        n.discovering = true
        n.discover()
        n.initState()
        n.fetchBlocks()
    }
    go func() {
        for {
            log.Println("node start")
            isM, isL := store.MoveToNextMiner();
            if isL {
                n.runAsLeader()
            } else {
                n.wg = concurrent.NewCountDownLatch(1)
                //n.wg.Add(1)
                n.prepLock.Lock()
                n.prep = true
                n.prepCond.Broadcast()
                n.prepLock.Unlock()
                if isM {
                    n.runAsMember()
                }
                //n.wg.Wait()
                if !n.wg.WaitTimeout(5 * time.Second) { // If not, next will be nil member
                   fmt.Println("Offline")
                   return
                }
            }
            log.Println("node stop")
            time.Sleep(5 * time.Second)
            log.Println("Current height ", store.GetCurrentBlockHeight())
            // todo if timeout still can go. why
        }
    }()
}

func (n *Node) runAsLeader() {
    leader1 := consensus.NewLeader(n.prvKey.PubKey, store.GetMiners())
    store.SetLeader(leader1)
    block := store.GenerateBlock()
    log.Println("node leader is preparing process consensus for round 1")
    if msg, err := json.Marshal(block); err ==nil {
        log.Println("node leader is going to process consensus for round 1")
        sig, bitmap := leader1.ProcessConsensus(msg)
        multiSig := &bean.MultiSignature{Sig: sig, Bitmap: bitmap}
        log.Println("node leader is preparing process consensus for round 2")
        if msg, err := json.Marshal(multiSig); err == nil {
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
    if _, ps := network.SendMessage(peers, block); len(ps) > 0 {
        fmt.Println("Offline peers: ", ps)
        store.RemovePeers(ps)
    }
}

func (n *Node) runAsMember() {
    member1 := consensus.NewMember(store.GetLeader(), store.GetPrvKey())
    store.SetMember(member1)
    log.Println("node member is going to process consensus for round 1")
    bytes := member1.ProcessConsensus()
    log.Println("node member finishes consensus for round 1")
    block := &bean.Block{}
    //n.wg = &sync.WaitGroup{}
    //n.wg.Add(1)
    if json.Unmarshal(bytes, block) == nil {
        member2 := consensus.NewMember(store.GetLeader(), store.GetPrvKey())
        store.SetMember(member2)
        log.Println("node member is going to process consensus for round 2")
        member2.ProcessConsensus()
        log.Println("node member finishes consensus for round 2")
    }
    //log.Println("node member is going to wait")
    //n.wg.WaitTimeout()
    //log.Println("node member finishes wait")
}

func (n *Node) ProcessBlock(block *bean.Block, del bool) {
    if del {
        n.prepLock.Lock()
        for !n.prep {
            n.prepCond.Wait()
        }
        n.prep = false
        n.prepLock.Unlock()
    }
    if fee := n.processBlock(block, del); fee == nil {
        fmt.Println("Offline. start to fetch block")
        n.fetchBlocks()
    }
    // todo receive two, should not !!! the same goes with other similar cases
    // todo maybe receive two consecutive blocks
    if del {
        n.wg.Done()
    }
}

func (n *Node) processBlock(block *bean.Block, del bool) *big.Int {
    log.Println("node receive block", *block)
    fmt.Println("Process block leader = ", bean.Addr(block.Header.LeaderPubKey), " height = ", block.Header.Height)
    return store.ExecuteTransactions(block, del)
}

func (n *Node) discover() {
    fmt.Println("discovering 1")
    // todo
    ips := network.GetIps()
    if len(ips) == 0 {
        fmt.Println("Error")
    } else if len(ips) > 1 {
        fmt.Println("Strange")
    }
    var msg *bean.PeerInfo
    if store.LOCAL_TEST {
        msg = &bean.PeerInfo{Pk: n.prvKey.PubKey, Ip: "127.0.0.1", Port: 55557}
    } else {
        msg = &bean.PeerInfo{Pk: n.prvKey.PubKey, Ip: ips[0], Port: 55555}
    }
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
    fmt.Println("user starting process a newcomer")
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
    fmt.Println("ProcessNewPeer ", *peerList, peers, newcomer)
    network.SendMessage([]*network.Peer{newPeer}, peerList)
    network.SendMessage(peers, &bean.PeerInfoList{List:[]*bean.PeerInfo{newcomer}})
}

func (n *Node) ProcessPeerList(list *bean.PeerInfoList) {
    fmt.Println("discovering 5 ", *list)
    for _, t := range list.List {
        store.AddPeer(&network.Peer{IP:network.IP(t.Ip), Port:network.Port(t.Port), PubKey:t.Pk})
    }
    fmt.Println("discovering 6")
    if n.discovering {
        n.discoverWg.Done()
    }
    fmt.Println("discovering 7")
}

func (n *Node) fetchBlocks() {
    //peers := store.GetPeers()
    //if len(peers) == 0 {
    //    log.Errorf("Fuck")
    //    return
    //}
    n.fetchCond = sync.NewCond(&n.fetchLock)
    n.fetchLock.Lock()
    defer n.fetchLock.Unlock()
    n.curMaxHeight = 2<<60
    req := &bean.BlockReq{Height:store.GetCurrentBlockHeight(), Pk:store.GetPubKey()}
    //network.SendMessage([]*network.Peer{peers[0]}, req)
    network.SendMessage([]*network.Peer{store.Admin}, req)
    fmt.Println("fetching 1")
    for n.curMaxHeight != store.GetCurrentBlockHeight() {
        fmt.Println("fetching 2: ", n.curMaxHeight, store.GetCurrentBlockHeight())
        n.fetchCond.Wait()
        fmt.Println("fetching 3: ", n.curMaxHeight, store.GetCurrentBlockHeight())
    }
}

func (n *Node) ProcessBlockResp(resp *bean.BlockResp) {
    fmt.Println("fetching 4")
    for _, b := range resp.Blocks {
        n.processBlock(b, false)
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
    fmt.Println("pk = ", req.Pk)
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

func (n *Node) ReportOfflinePeers(peers []*network.Peer) {
    msg := make([]*bean.PeerInfo, len(peers))
    for i, p := range peers {
        msg[i] = &bean.PeerInfo{Pk:p.PubKey, Ip:string(p.IP), Port:int32(p.Port)}
    }
    network.SendMessage([]*network.Peer{store.Admin}, msg)
}

func (n *Node) Ping(peer *network.Peer) error {
    addr := bean.Addr(peer.PubKey)
    if latch, exist := n.pingLatches[addr]; exist {
        return &util.DupOpError{}
    } else {
        latch = concurrent.NewCountDownLatch(1)
        n.pingLatches[addr] = latch
        ping := &bean.Ping{Pk:store.GetPubKey()}
        network.SendMessage([]*network.Peer{peer}, ping)
        if latch.WaitTimeout(5 * time.Second) {
            return &util.TimeoutError{}
        } else {
            return nil
        }
    }
}

func (n *Node) ProcessPing(peer *network.Peer, ping *bean.Ping)  {
    network.SendMessage([]*network.Peer{peer}, &bean.Pong{Pk:store.GetPubKey()})
}

func (n *Node) ProcessPong(peer *network.Peer, ping *bean.Ping) {
    addr := bean.Addr(peer.PubKey)
    if latch, exist := n.pingLatches[addr]; exist {
        latch.Done()
        delete(n.pingLatches, addr)
    }
}

func (n *Node) ProcessOfflinePeers(peers []*bean.PeerInfo)  {
    if !store.IsAdmin() {
        log.Errorf("I am not admin but receive offline peers.")
        return
    }
    for _, p := range peers {
        store.RemovePeer(&network.Peer{PubKey:p.Pk, IP:network.IP(p.Ip), Port:network.Port(p.Port)})
    }
}

func (n *Node) initState() {
    bs := database.LoadAllBlock(0)
    for _, b := range bs {
        store.ExecuteTransactions(b, false)
    }
}
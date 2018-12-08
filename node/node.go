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
    "BlockChainTest/pool"
    "BlockChainTest/accounts"
)

var (
    once sync.Once
    node *Node
)

type Node struct {
    prvKey *mycrypto.PrivateKey
    curMaxHeight int64
    pingLatches map[accounts.CommonAddress]concurrent.CountDownLatch
}

func newNode(prvKey *mycrypto.PrivateKey) *Node {
    n := &Node{prvKey: prvKey}
    //n.prepCond = sync.NewCond(&n.prepLock)
    n.pingLatches = make(map[accounts.CommonAddress]concurrent.CountDownLatch)
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
        //n.discovering = false
    } else {
        //n.discovering = true
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
                //n.prepLock.Lock()
                //n.prep = true
                //n.prepCond.Broadcast()
                //n.prepLock.Unlock()
                if isM {
                    n.runAsMember()
                }
                //n.wg.Wait()
                if block := pool.ObtainOne(func(msg interface{}) bool {
                    _, ok := msg.(*bean.Block)
                    return ok
                }, 5 * time.Second); block != nil {
                    if b, ok := block.(*bean.Block); ok {
                        n.processBlock(b)
                    }
                } else {
                   fmt.Println("Offline")
                   return
                }
            }
            log.Println("node stop")
            time.Sleep(10 * time.Second)
            log.Println("Current height ", database.GetMaxHeight())
            // todo if timeout still can go. why
        }
    }()
}

func (n *Node) runAsLeader() {
    leader1 := consensus.NewLeader(n.prvKey.PubKey, store.GetMiners())
    block, _ := store.GenerateBlock()
    log.Println("node leader is preparing process consensus for round 1")
    if msg, err := json.Marshal(block); err ==nil {
        log.Println("node leader is going to process consensus for round 1")
        err, sig, bitmap := leader1.ProcessConsensus(msg)
        if err != nil {
            fmt.Println("Error occurs", err)
            panic(err)
        }
        multiSig := &bean.MultiSignature{Sig: sig, Bitmap: bitmap}
        log.Println("node leader is preparing process consensus for round 2")
        if msg, err := json.Marshal(multiSig); err == nil {
            leader2 := consensus.NewLeader(n.prvKey.PubKey, store.GetMiners())
            log.Println("node leader is going to process consensus for round 2")
            leader2.ProcessConsensus(msg)
            log.Println("node leader finishes process consensus for round 2")
            log.Println("node leader is going to send block")
            block.MultiSig = multiSig
            n.ProcessBlock(block) // process before sending
            n.sendBlock(block)
            log.Println("node leader finishes sending block")
        }
    }
}

func (n *Node) sendBlock(block *bean.Block) {
    peers := store.GetPeers()
    //todo concurrent
    if _, ps := network.SendMessage(peers, block); len(ps) > 0 {
        fmt.Println("Offline peers: ", ps)
        //store.RemovePeers(ps)
    }
}

func (n *Node) runAsMember() {
    member1 := consensus.NewMember(store.GetLeader(), store.GetPrvKey())
    log.Println("node member is going to process consensus for round 1")
    bytes := member1.ProcessConsensus(func(setup *bean.Setup) bool {
        block := &bean.Block{}
        if err := json.Unmarshal(setup.Msg, block); err == nil {
            //TODO block
            return true
        } else {
            return false
        }
    })
    log.Println("node member finishes consensus for round 1")
    block := &bean.Block{}
    //n.wg = &sync.WaitGroup{}
    //n.wg.Add(1)
    if json.Unmarshal(bytes, block) == nil {
        member2 := consensus.NewMember(store.GetLeader(), store.GetPrvKey())
        log.Println("node member is going to process consensus for round 2")
        member2.ProcessConsensus(func(setup *bean.Setup) bool {
            return true
        })
        log.Println("node member finishes consensus for round 2")
    }
    //log.Println("node member is going to wait")
    //n.wg.WaitTimeout()
    //log.Println("node member finishes wait")
}

func (n *Node) ProcessBlock(block *bean.Block) {
    //if del {
    //    n.prepLock.Lock()
    //    for !n.prep {
    //        n.prepCond.Wait()
    //    }
    //    n.prep = false
    //    n.prepLock.Unlock()
    //}
    if fee := n.processBlock(block); fee == nil {
        fmt.Println("Offline. start to fetch block")
        n.fetchBlocks()
    }
    // todo receive two, should not !!! the same goes with other similar cases
    // todo maybe receive two consecutive blocks
    //if del {
    //    n.wg.Done()
    //}
}

func (n *Node) processBlock(block *bean.Block) *big.Int {
    log.Println("node receive block", *block)
    fmt.Println("Process block leader = ", accounts.PubKey2Address(block.Header.LeaderPubKey).Hex(), " height = ", block.Header.Height)
    return store.ExecuteTransactions(block)
}

func (n *Node) discover() bool {
    fmt.Println("discovering 1")
    // todo
    ips := network.GetIps()
    if len(ips) == 0 {
        fmt.Println("Error")
    } else if len(ips) > 1 {
        fmt.Println("Strange")
    }
    var msg *bean.PeerInfo
    if store.LocalTest {
        msg = &bean.PeerInfo{Pk: n.prvKey.PubKey, Ip: "127.0.0.1", Port: 55557}
    } else {
        msg = &bean.PeerInfo{Pk: n.prvKey.PubKey, Ip: ips[0], Port: 55555}
    }
    peers := []*bean.Peer{store.Admin}
    network.SendMessage(peers, msg)
    fmt.Println("discovering 3")
    if msg := pool.ObtainOne(func(msg interface{}) bool {
        _, ok := msg.(*bean.FirstPeerInfoList)
        return ok
    }, 5 * time.Second); msg != nil {
        if pil, ok := msg.(*bean.FirstPeerInfoList); ok {
            for _, t := range pil.List {
                store.AddPeer(&bean.Peer{IP:bean.IP(t.Ip), Port:bean.Port(t.Port), PubKey:t.Pk})
            }
        }
        return true
    } else {
        fmt.Println("Cannot get peers")
        return false
    }
}

func (n *Node) ProcessNewPeer(newcomer *bean.PeerInfo) {
    fmt.Println("user starting process a newcomer")
    peers := store.GetPeers()
    newPeer := &bean.Peer{
        IP:bean.IP(newcomer.Ip),
        Port: bean.Port(newcomer.Port),
        PubKey: newcomer.Pk}
    store.AddPeer(newPeer)
    list := make([]*bean.PeerInfo, 0)
    for _, p := range peers {
        t := &bean.PeerInfo{Pk: p.PubKey, Ip:string(p.IP), Port:int32(p.Port)}
        list = append(list, t)
    }
    peerList := &bean.PeerInfoList{List:list}
    fmt.Println("ProcessNewPeer ", *peerList, peers, newcomer)
    network.SendMessage([]*bean.Peer{newPeer}, peerList)
    network.SendMessage(peers, &bean.FirstPeerInfoList{List:[]*bean.PeerInfo{newcomer}})
}

func (n *Node) ProcessPeerList(list *bean.PeerInfoList) {
    fmt.Println("discovering 5 ", *list)
    for _, t := range list.List {
        store.AddPeer(&bean.Peer{IP:bean.IP(t.Ip), Port:bean.Port(t.Port), PubKey:t.Pk})
    }
    //if n.discovering {
    //    n.discoverWg.Done()
    //}
}

func (n *Node) fetchBlocks() {
    n.curMaxHeight = 2<<60
    req := &bean.BlockReq{Height:database.GetMaxHeight(), Pk:store.GetPubKey()}
    //network.SendMessage([]*bean.Peer{peers[0]}, req)
    network.SendMessage([]*bean.Peer{store.Admin}, req)
    fmt.Println("fetching 1")
    for n.curMaxHeight != database.GetMaxHeight() {
       fmt.Println("fetching 2: ", n.curMaxHeight, database.GetMaxHeight())
       if msg := pool.ObtainOne(func(msg interface{}) bool {
           if block, ok := msg.(*bean.Block); ok {
               return block != nil && block.Header != nil && block.Header.Height == database.GetMaxHeight() + 1
           } else {
               return false
           }
       }, 5 * time.Second); msg != nil {
           if block, ok := msg.(*bean.Block); ok {
               n.processBlock(block)
           }
       }
       fmt.Println("fetching 3: ", n.curMaxHeight, database.GetMaxHeight())
   }
}

func (n *Node) ProcessBlockReq(req *bean.BlockReq) {
    from := req.Height + 1
    size := int64(2)
    fmt.Println("pk = ", req.Pk)
    peers := []*bean.Peer{store.GetPeer(req.Pk)}
    fmt.Println("ProcessBlockReq")
    for i := from; i <= database.GetMaxHeight(); {
        fmt.Println("ProcessBlockReq 1 ", i)
        bs := database.GetBlocksFrom(i, size)
        resp := &bean.BlockResp{Height:database.GetMaxHeight(), Blocks:bs}
        network.SendMessage(peers, resp)
        i += int64(len(bs))
        fmt.Println("ProcessBlockReq 2 ", i)
    }
}

func (n *Node) ReportOfflinePeers(peers []*bean.Peer) {
    msg := make([]*bean.PeerInfo, len(peers))
    for i, p := range peers {
        msg[i] = &bean.PeerInfo{Pk:p.PubKey, Ip:string(p.IP), Port:int32(p.Port)}
    }
    network.SendMessage([]*bean.Peer{store.Admin}, msg)
}

func (n *Node) Ping(peer *bean.Peer) error {
    addr := accounts.PubKey2Address(peer.PubKey)
    if latch, exist := n.pingLatches[addr]; exist {
        return &util.DupOpError{}
    } else {
        latch = concurrent.NewCountDownLatch(1)
        n.pingLatches[addr] = latch
        ping := &bean.Ping{Pk:store.GetPubKey()}
        network.SendMessage([]*bean.Peer{peer}, ping)
        if latch.WaitTimeout(5 * time.Second) {
            return &util.TimeoutError{}
        } else {
            return nil
        }
    }
}

func (n *Node) ProcessPing(peer *bean.Peer, ping *bean.Ping)  {
    network.SendMessage([]*bean.Peer{peer}, &bean.Pong{Pk:store.GetPubKey()})
}

func (n *Node) ProcessPong(peer *bean.Peer, ping *bean.Ping) {
    addr := accounts.PubKey2Address(peer.PubKey)
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
        store.RemovePeer(&bean.Peer{PubKey:p.Pk, IP:bean.IP(p.Ip), Port:bean.Port(p.Port)})
    }
}

func (n *Node) initState() {
    bs := database.GetAllBlocks()
    for _, b := range bs {
        store.ExecuteTransactions(b)
    }
}
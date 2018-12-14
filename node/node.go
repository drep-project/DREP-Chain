package node

import (
    "BlockChainTest/store"
    "BlockChainTest/bean"
    "BlockChainTest/consensus"
    "sync"
    "strconv"
    "BlockChainTest/log"
    "BlockChainTest/network"
    "BlockChainTest/mycrypto"
    "time"
    "math/big"
    "BlockChainTest/config"
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

func (n *Node) Start(config *config.NodeConfig) {
    if config.Boot {
        //n.discovering = false
    } else {
        //n.discovering = true
        n.discover()
        n.initState()
        n.fetchBlocks()
    }
    go func() {
        for {
            log.Trace("node start")
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
                if block := pool.ObtainOneMsg(func(msg interface{}) bool {
                    _, ok := msg.(*bean.Block)
                    return ok
                }, 5 * time.Second); block != nil {
                    if b, ok := block.(*bean.Block); ok {
                        n.processBlock(b)
                    }
                } else {
                   log.Error("Offline")
                   return
                }
            }
            log.Trace("node stop")
            time.Sleep(10 * time.Second)
            log.Debug("Current height ", database.GetMaxHeight())
            // todo if timeout still can go. why
        }
    }()
}

func (n *Node) runAsLeader() {
    leader1 := consensus.NewLeader(n.prvKey.PubKey, store.GetMiners())
    block, _ := store.GenerateBlock(leader1.GetMembers())
    log.Trace("node leader is preparing process consensus for round 1", "Block",block)
    if msg, err := json.Marshal(block); err ==nil {
        log.Trace("node leader is going to process consensus for round 1")
        err, sig, bitmap := leader1.ProcessConsensus(msg)
        if err != nil {
            var str = err.Error()
            log.Error("Error occurs","msg", str)
            panic(err)
        }
        multiSig := &bean.MultiSignature{Sig: sig, Bitmap: bitmap}
        log.Trace("node leader is preparing process consensus for round 2")
        if msg, err := json.Marshal(multiSig); err == nil {
            leader2 := consensus.NewLeader(n.prvKey.PubKey, store.GetMiners())
            log.Trace("node leader is going to process consensus for round 2")
            leader2.ProcessConsensus(msg)
            log.Trace("node leader finishes process consensus for round 2")
            log.Trace("node leader is going to send block")
            block.MultiSig = multiSig
            n.ProcessBlock(block) // process before sending
            n.sendBlock(block)
            log.Trace("node leader finishes sending block")
        }
    }
}

func (n *Node) sendBlock(block *bean.Block) {
    peers := store.GetPeers()
    //todo concurrent
    if _, ps := network.SendMessage(peers, block); len(ps) > 0 {
        log.Trace("Offline peers: ", ps)
        //store.RemovePeers(ps)
    }
}

func (n *Node) runAsMember() {
    member1 := consensus.NewMember(store.GetLeader(), store.GetPrvKey())
    log.Trace("node member is going to process consensus for round 1")
    bytes := member1.ProcessConsensus(func(setup *bean.Setup) bool {
        block := &bean.Block{}
        if err := json.Unmarshal(setup.Msg, block); err == nil {
            //TODO block
            return true
        } else {
            return false
        }
    })
    log.Trace("node member finishes consensus for round 1")
    block := &bean.Block{}
    //n.wg = &sync.WaitGroup{}
    //n.wg.Add(1)
    if json.Unmarshal(bytes, block) == nil {
        member2 := consensus.NewMember(store.GetLeader(), store.GetPrvKey())
        log.Trace("node member is going to process consensus for round 2")
        member2.ProcessConsensus(func(setup *bean.Setup) bool {
            return true
        })
        log.Trace("node member finishes consensus for round 2")
    }
    //log.Trace("node member is going to wait")
    //n.wg.WaitTimeout()
    //log.Trace("node member finishes wait")
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
        log.Trace("Offline. start to fetch block")
        n.fetchBlocks()
    }
    // todo receive two, should not !!! the same goes with other similar cases
    // todo maybe receive two consecutive blocks
    //if del {
    //    n.wg.Done()
    //}
}

func (n *Node) processBlock(block *bean.Block) *big.Int {
    log.Trace("Process block leader.", "LeaderPubKey", accounts.PubKey2Address(block.Header.LeaderPubKey).Hex(), " height ", strconv.FormatInt(block.Header.Height,10))
    return store.ExecuteTransactions(block)
}

func (n *Node) discover() bool {
    log.Trace("discovering 1")
    // todo
    ips := network.GetIps()
    if len(ips) == 0 {
        log.Error("Error")
    } else if len(ips) > 1 {
        log.Trace("Strange")
    }
    var msg *bean.PeerInfo
    if store.LocalTest {
        msg = &bean.PeerInfo{Pk: n.prvKey.PubKey, Ip: "127.0.0.1", Port: 55557}
    } else {
        msg = &bean.PeerInfo{Pk: n.prvKey.PubKey, Ip: ips[0], Port: 55555}
    }
    peers := []*bean.Peer{store.Admin}
    network.SendMessage(peers, msg)
    log.Trace("discovering 3")
    if msg := pool.ObtainOneMsg(func(msg interface{}) bool {
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
        log.Trace("Cannot get peers")
        return false
    }
}

func (n *Node) ProcessNewPeer(newcomer *bean.PeerInfo) {
    log.Trace("user starting process a newcomer")
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
    log.Trace("ProcessNewPeer ", *peerList, peers, newcomer)
    network.SendMessage([]*bean.Peer{newPeer}, peerList)
    network.SendMessage(peers, &bean.FirstPeerInfoList{List:[]*bean.PeerInfo{newcomer}})
}

func (n *Node) ProcessPeerList(list *bean.PeerInfoList) {
    log.Trace("discovering 5 ", *list)
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
    //network.SendMessage([]*network.Peer{peers[0]}, req)
    network.SendMessage([]*bean.Peer{store.Admin}, req)
    log.Trace("fetching 1")
    for n.curMaxHeight != database.GetMaxHeight() {
       log.Trace("fetching 2: ", n.curMaxHeight, database.GetMaxHeight())
       if msg := pool.ObtainOneMsg(func(msg interface{}) bool {
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
       log.Trace("fetching 3: ", n.curMaxHeight, database.GetMaxHeight())
   }
}

func (n *Node) ProcessBlockReq(req *bean.BlockReq) {
    from := req.Height + 1
    size := int64(2)
    log.Trace("pk = ", req.Pk)
    peers := []*bean.Peer{store.GetPeer(req.Pk)}
    log.Trace("ProcessBlockReq")
    for i := from; i <= database.GetMaxHeight(); {
        log.Trace("ProcessBlockReq 1 ", i)
        bs := database.GetBlocksFrom(i, size)
        resp := &bean.BlockResp{Height:database.GetMaxHeight(), Blocks:bs}
        network.SendMessage(peers, resp)
        i += int64(len(bs))
        log.Trace("ProcessBlockReq 2 ", i)
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
        log.Error("I am not admin but receive offline peers.")
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
package role

import (
    "sync"
    "BlockChainTest/network"
    "BlockChainTest/crypto"
    "BlockChainTest/bean"
    "BlockChainTest/log"
    "fmt"
)

const (
    waiting             = 0
    done                = 1
)

type Newcomer struct {
    address *bean.Address
    prvKey *crypto.PrivateKey
    neighbour *network.Peer
    state int
    wg *sync.WaitGroup
}

func NewJoiner(peer *network.Peer) *Newcomer {
    sk,pk,error := crypto.GetRandomKQ()
    if error != nil {
        log.Println("generate key error:", error)
        return nil
    }
    newcomer := &Newcomer{}
    newcomer.prvKey = &crypto.PrivateKey{Prv: sk, PubKey: pk}

    address := bean.Addr(pk)
    newcomer.address = &address

    newcomer.state = waiting

    newcomer.neighbour = peer

    newcomer.wg = &sync.WaitGroup{}
    newcomer.wg.Add(1)

    return  newcomer
}

// run this func at first time
func (n *Newcomer) ProcessJoin()  {
    msg := &bean.Newcomer{}
    msg.Pk = n.prvKey.PubKey
    msg.Ip = "192.168.3.113"
    msg.Port = 55555
    log.Println("there is a newcomer request to join the blockchain family!")
    log.Println("start request.")

    var peers = []*network.Peer{n.neighbour}
    network.SendMessage(peers, msg)
    log.Println("n.neighbour: ", n.neighbour)
    n.wg.Wait()
}

func (n *Newcomer) ProcessWelcome(list *bean.ListOfPeer) {
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


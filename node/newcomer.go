package node

import (
    "sync"
    "BlockChainTest/network"
    "BlockChainTest/crypto"
    "BlockChainTest/bean"
    "BlockChainTest/store"
    "BlockChainTest/log"
)

const (
    waiting             = 0
    done                = 1
)

type Newcomer struct {
    Node
    neighbour *network.Peer
    state int
}

func NewJoiner() *Newcomer {
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
    // TODO: a hard coding server
    peers := store.GetPeers()
    newcomer.neighbour = peers[0]

    newcomer.wg = &sync.WaitGroup{}
    newcomer.wg.Add(1)

    return  newcomer
}

// run this func at first time
func (n *Newcomer) ProcessJoin()  {
    msg := &bean.Newcomer{}
    msg.Pk = n.prvKey.PubKey

    log.Println("there is a newcomer request to join the blockchain family!")
    log.Println("start request.")

    var peers = []*network.Peer{n.neighbour}
    network.SendMessage(peers, msg)
    n.wg.Wait()
}

func (n *Newcomer) ProcessWelcome(list *bean.ListOfPeer) {
    log.Println("welcome newcomer! it's done.")
    peerStore := network.GetStore()

    // store the peers in the local memory.
    for _, item := range list.List {
        pubKey := item.Pk
        address := bean.Addr(pubKey)
        peer := &network.Peer{}
        peer.PubKey = pubKey
        peer.Address = address
        peerStore.Store[address] = peer
    }

    n.state = done
    n.wg.Done()
}


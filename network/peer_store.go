package network

import (
    "sync"
    "BlockChainTest/bean"
)

var sharedPeerStore *PeerStore
var once sync.Once

type PeerStore struct {
    Store map[bean.Address]*Peer
}


// Returns the singleton PeerStore instance.
func GetStore () *PeerStore {
    once.Do(func() {
        sharedPeerStore = &PeerStore{}
        sharedPeerStore.Store = make(map[bean.Address]*Peer)
    })
    return sharedPeerStore
}

// Adds a Peer to the map table.
func (ps *PeerStore) AddPeer(peer *Peer) {
    pk := bean.Addr(peer.PubKey)
    ps.Store[pk] = peer
}

// Returns the current number of peers in the table.
func (ps *PeerStore) GetPeerCount () int {
    return len(ps.Store)
}



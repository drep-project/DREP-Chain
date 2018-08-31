package network

import "sync"

var sharedPeerStore *PeerStore
var once sync.Once

type PeerStore struct {
    store map[int]string
}

// Returns the singleton PeerStore instance.
func (ps *PeerStore) GetStore () *PeerStore {
    once.Do(func() {
        sharedPeerStore = &PeerStore{}
    })
    return sharedPeerStore

}

// Adds a Peer to the table.
func (ps *PeerStore) AddPeerPair (pk string, peer *Peer) {

}

// Returns the number of peers in the table.
func (ps *PeerStore) GetPeerCount () int {
    return len(ps.store)
}



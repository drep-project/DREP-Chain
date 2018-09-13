package store

import (
    "BlockChainTest/bean"
    "BlockChainTest/network"
)

var (
    store map[bean.Address]*network.Peer
)

func AddPeer(peer *network.Peer) {
    addr := bean.Addr(peer.PubKey)
    if _, exists := store[addr]; !exists {
        store[addr] = peer
    }
}

func GetPeers() []*network.Peer {
    result := make([]*network.Peer, 0)
    for _, v := range store {
        if !v.PubKey.Equal(pubKey) {
            result = append(result, v)
        }
    }
    return result
}
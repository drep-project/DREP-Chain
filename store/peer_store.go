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
    store[addr] = peer
}

func GetPeers() []*network.Peer {
    result := make([]*network.Peer, 0)
    for _, v := range peers {
        if !v.PubKey.Equal(pubKey) {
            result = append(result, v)
        }
    }
    return result
}
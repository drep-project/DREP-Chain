package store

import (
    "BlockChainTest/bean"
    "BlockChainTest/network"
    "BlockChainTest/log"
    "BlockChainTest/mycrypto"
)

var (
    store = make(map[bean.Address]*network.Peer)
    miners []*network.Peer
    nextMiners []*network.Peer
    nextBlockHeight int64
    currentMinerIndex int
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


func MoveToNextMiner() bool {
    lock.Lock()
    currentMinerIndex = (currentMinerIndex + 1) % len(miners)
    log.Println("Current miner index:", currentMinerIndex, " my index: ", myIndex)
    r := currentMinerIndex == myIndex
    leader = nil
    member = nil
    lock.Unlock()
    return r
}

func GetLeader() *network.Peer {
    return miners[currentMinerIndex]
}


func GetMiners() []*network.Peer {
    return miners
}

func AddMiner(miner *network.Peer, height int64) {
    if miner != nil {
        if len(miners) < maxMinerNumber {
            miners = append(miners, miner)
        } else {
            miners = append(miners[1:], miner)
        }
    }
    for _, m := range miners {
        if m.PubKey.Equal(pubKey) {
            role = bean.MINER
            return
        }
    }
    role = bean.OTHER
}

func ContainsMiner(pubKey *mycrypto.Point) bool {
    for _, v:= range miners {
        if v.PubKey.Equal(pubKey) {
            return true
        }
    }
    return false
}

func GetMiner(pubKey *mycrypto.Point) *network.Peer {
    for _, v:= range miners {
        if v.PubKey.Equal(pubKey) {
            return v
        }
    }
    return nil
}

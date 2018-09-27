package store

import (
    "BlockChainTest/bean"
    "BlockChainTest/network"
    "BlockChainTest/mycrypto"
)

var (
    peers                    = make(map[bean.Address]*network.Peer)
    curMiners   []*network.Peer
    miners                  = make([]*network.Peer, 0)
    curMiner    int
    minerIndex  int
    adminPubKey *mycrypto.Point
)

func AddPeer(peer *network.Peer) {
    addr := bean.Addr(peer.PubKey)
    if _, exists := peers[addr]; !exists {
        peers[addr] = peer
    }
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


func MoveToNextMiner() (bool, bool) {
    lock.Lock()
    defer lock.Unlock()
    SetLeader(nil)
    SetMember(nil)
    curMiner++
    if curMiner == len(curMiners) {
        if minerIndex < len(miners) - 1 {
            minerIndex++
            curMiners = append(curMiners[1:], miners[minerIndex])
        }
        curMiner = 0
    }
    isM := false
    for _, m := range curMiners {
        if m.PubKey.Equal(GetPubKey()) {
            isM = true
            break
        }
    }
    return isM, curMiners[curMiner].PubKey.Equal(GetPubKey())
}

func GetLeader() *network.Peer {
    return curMiners[curMiner]
}


func GetMiners() []*network.Peer {
    return curMiners
}

func AddMiner(addr bean.Address) {
    a := string(addr)
    for _, p := range peers {
        if string(bean.Addr(p.PubKey)) == a {
            miners = append(miners, p)
        }
    }
}

func GetAdminPubKey() *mycrypto.Point {
    return adminPubKey
}

func GetPeer(pk *mycrypto.Point) *network.Peer {
    for _, p := range peers {
        if p.PubKey.Equal(pk) {
            return p
        }
    }
    return nil
}
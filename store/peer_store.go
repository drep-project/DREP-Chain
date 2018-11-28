package store

import (
    "BlockChainTest/network"
    "BlockChainTest/mycrypto"
    "BlockChainTest/accounts"
)

var (
    peers                    = make(map[accounts.CommonAddress]*network.Peer)
    curMiners   []*network.Peer
    miners                  = make([]*network.Peer, 0)
    curMiner    int
    minerIndex  int
    adminPubKey *mycrypto.Point
)

func AddPeer(peer *network.Peer) {
    addr := accounts.PubKey2Address(peer.PubKey)
    if _, exists := peers[addr]; !exists {
        peers[addr] = peer
    }
}

func RemovePeer(peer *network.Peer) {
    addr := accounts.PubKey2Address(peer.PubKey)
    delete(peers, addr)
}

func RemovePeers(peers []*network.Peer) {
    for _, p := range peers {
        RemovePeer(p)
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

func AddMiner(addr accounts.CommonAddress) {
    a := addr.Hex()
    for _, p := range peers {
        if accounts.PubKey2Address(p.PubKey).Hex() == a {
            miners = append(miners, p)
        }
    }
}

func GetAdminPubKey() *mycrypto.Point {
    return adminPubKey
}

func IsAdmin() bool {
    return adminPubKey.Equal(pubKey)
}

func GetPeer(pk *mycrypto.Point) *network.Peer {
    for _, p := range peers {
        if p.PubKey.Equal(pk) {
            return p
        }
    }
    return nil
}
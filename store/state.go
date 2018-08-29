package store

import (
    "sync"
    "BlockChainTest/network"
    "BlockChainTest/bean"
    "BlockChainTest/consensus"
    "BlockChainTest/crypto"
)


const index = 0

var (
    role int
    leader *consensus.Leader
    member *consensus.Member
    miningState int
    miners []*network.Peer
    minerIndex map[bean.Address]int
    peers []*network.Peer
    blockHeight int
    block *bean.Block
    lock sync.Locker
    prvKey *bean.PrivateKey
    pubKey *bean.Point
    address bean.Address
    //k0, k1, k2 []byte
    //pub0, pub1, pub2 *bean.Point
    //prv0, prv1, prv2 *bean.PrivateKey
    //miner0, miner1, miner2 *network.Peer
    //ip0, ip1, ip2 network.IP
    //port0, port1, port2 network.Port
    //peer0, peer1, peer2 *network.Peer
)

func init()  {
    lock = &sync.Mutex{}
    //prvKey, _ = crypto.GetPrivateKey()
    //pubKey = GetPubKey()
    curve := crypto.GetCurve()
    k0 := []byte{0x22, 0x11}
    k1 := []byte{0x14, 0x44}
    k2 := []byte{0x11, 0x55}
    pub0 := curve.ScalarBaseMultiply(k0)
    pub1 := curve.ScalarBaseMultiply(k1)
    pub2 := curve.ScalarBaseMultiply(k2)
    prv0 := &bean.PrivateKey{Prv: k0, PubKey: pub0}
    prv1 := &bean.PrivateKey{Prv: k1, PubKey: pub1}
    prv2 := &bean.PrivateKey{Prv: k2, PubKey: pub2}
    ip0 := network.IP("192.168.3.13")
    ip1 := network.IP("192.168.3.43")
    ip2 := network.IP("192.168.3.73")
    port0 := network.Port(55555)
    port1 := network.Port(55555)
    port2 := network.Port(55555)
    peer0 := &network.Peer{IP: ip0, Port: port0, PubKey: pub0}
    peer1 := &network.Peer{IP: ip1, Port: port1, PubKey: pub1}
    peer2 := &network.Peer{IP: ip2, Port: port2, PubKey: pub2}
    miners = []*network.Peer{peer0, peer1, peer2}
    peers = []*network.Peer{peer0, peer1, peer2}
    switch index {
    case 0:
        pubKey = pub0
        prvKey = prv0
        leader = consensus.NewLeader(pub0, peers)
        member = nil
    case 1:
        pubKey = pub1
        prvKey = prv1
        leader = nil
        member = consensus.NewMember(peer0, prvKey)
    case 2:
        pubKey = pub2
        prvKey = prv2
        leader = nil
        member = consensus.NewMember(peer0, prvKey)
    }
}

func ChangeRole(r int) {
    lock.Lock()
    role = r
    miners = GetMiners()
    lock.Unlock()
}

func GetRole() int {
    return role
}

func ChangeMiningState(s int) {
    lock.Lock()
    miningState = s
    lock.Unlock()
}

func GetMiningState() int {
    return miningState
}

func GetMiners() []*network.Peer {
    return miners
}

func ContainsMiner(pubKey *bean.Point) bool {
    for _, v:= range miners {
        if v.PubKey.Equal(pubKey) {
            return true
        }
    }
    return false
}

func GetMiner(pubKey *bean.Point) *network.Peer {
    for _, v:= range miners {
        if v.PubKey.Equal(pubKey) {
            return v
        }
    }
    return nil
}

func GetPeers() []*network.Peer {
    return peers
}

func GetBlockHeight() int {
    return blockHeight
}

func CheckState(r int, ms int) bool {
    return role == r && miningState == ms
}

func CheckRole(r int) bool {
    return role == r
}

func MoveToState(ms int) {
    miningState = ms
}

func GetLeader() *network.Peer {
    return nil
}

func GetBlock() *bean.Block {
    return &bean.Block{}
}

func SetBlock(b *bean.Block) {
    block = b
}

func GetPubKey() *bean.Point {
    return pubKey
}

func GetAddress() bean.Address {
    return address
}

func GetPrvKey() *bean.PrivateKey {
    return prvKey
}

func GetItSelfOnLeader() *consensus.Leader {
    return leader
}

func GetItSelfOnMember() *consensus.Member {
    return member
}

func GetMinerIndex(p *bean.Point) int {
    if pubKey.Equal(p) {
        return -1
    }
    if index, exist := minerIndex[p.Addr()]; exist {
        return index
    }
    return -1
}
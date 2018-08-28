package store

import (
    "sync"
    "BlockChainTest/network"
    "BlockChainTest/node"
    "BlockChainTest/bean"
    "BlockChainTest/consensus"
    "BlockChainTest/crypto"
)


var (
    role int
    leader *consensus.Leader
    member *consensus.Member
    miningState int
    miners []*network.Peer
    peers []*network.Peer
    blockHeight int
    block *bean.Block
    lock sync.Locker
    prvKey *bean.PrivateKey
    pubKey *bean.Point
    address bean.Address
    k0, k1 []byte
    pub0, pub1 *bean.Point
    prv0, prv1 *bean.PrivateKey
    miner0, miner1 *network.Peer
    ip0, ip1 network.IP
    port0, port1 network.Port
    peer0, peer1 *network.Peer
)

func init()  {
    lock = &sync.Mutex{}
    prvKey, _ = crypto.GetPrivateKey()
    pubKey = GetPubKey()
}

func ChangeRole(r int) {
    lock.Lock()
    role = r
    miners = GetMiners()
    if r == node.LEADER {
        leader = consensus.NewLeader(pub0, miners)
        member = nil
    } else {
        l := &node.Miner{PubKey: pub0, Peer: peer0}
        member = consensus.NewMember(l, prv1, pub1)
        leader = nil
    }
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
    curve := crypto.GetCurve()
    k0 = []byte{0x22, 0x11}
    k1 = []byte{0x14, 0x44}
    pub0 = curve.ScalarBaseMultiply(k0)
    pub1 = curve.ScalarBaseMultiply(k1)
    prv0 = &bean.PrivateKey{Prv: k0, PubKey: pub0}
    prv1 = &bean.PrivateKey{Prv: k1, PubKey: pub1}
    ip0 = network.IP("192.168.x.x")
    ip1 = network.IP("192.168.x.x")
    port0 = network.Port(1)
    port1 = network.Port(2)
    peer0 = &network.Peer{ip0, port0, pub0}
    peer1 = &network.Peer{ip1, port1, pub1}
    miner0 = &node.Miner{PubKey: pub0, Peer: peer0}
    miner1 = &node.Miner{PubKey: pub1, Peer: peer1}
    miners = make([]*node.Miner, 2)
    miners[0] = miner0
    miners[1] = miner1
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
    return block
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
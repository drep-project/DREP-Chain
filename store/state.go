package store

import (
    "sync"
    "BlockChainTest/network"
    "BlockChainTest/node"
    "BlockChainTest/bean"
    "BlockChainTest/consensus"
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
)

func init()  {

}

func ChangeRole(r int) {
    lock.Lock()
    role = r
    if r == node.LEADER {
        leader = consensus.NewLeader(pubKey, miners)
        member = nil
    } else {
        leader = nil
        member = consensus.NewMember(GetLeader(), prvKey, pubKey)
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
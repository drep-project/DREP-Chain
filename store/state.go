package store

import (
    "sync"
    "BlockChainTest/network"
    "BlockChainTest/common"
    "BlockChainTest/node"
    "bytes"
    "BlockChainTest/bean"
)


var (
    role int
    miningState int
    miners []*node.Miner
    peers []*network.Peer
    blockHeight int
    block *common.Block
    lock sync.Locker
    priKey *bean.Point
    pubKey *bean.Point
    address common.Address
)

func init()  {

}

func ChangeRole(r int) {
    lock.Lock()
    role = r
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

func GetMiners() []*node.Miner {
    return miners
}

func ContainsMiner(pubKey []byte) bool {
    for _, v:= range miners {
        if bytes.Equal(v.PubKey, pubKey) {
            return true
        }
    }
    return false
}

func GetMiner(pubKey []byte) *node.Miner {
    for _, v:= range miners {
        if bytes.Equal(v.PubKey, pubKey) {
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

func GetLeader() *node.Miner {
    return nil
}

func GetBlock() *common.Block {
    return block
}

func SetBlock(b *common.Block) {
    block = b
}

func GetPubKey() *bean.Point {
    return pubKey
}

func GetAddress() common.Address {
    return address
}

func GetPriKey() *bean.Point {
    return priKey
}
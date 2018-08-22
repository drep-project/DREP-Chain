package storage

import (
    "sync"
    "BlockChainTest/util"
    "BlockChainTest/network"
    "BlockChainTest/common"
)

const (
    LEADER     = 0
    MINER      = 1
    NON_MINER  = 2
)

const (

)

var (
    once sync.Once
    singleton *State
    lock sync.Locker
)

type Miner struct {
    PubKey []byte
    Address common.Address
    Peer *network.Peer
}

type State struct {
    role int
    miningState int
    miners []*Miner
    peers []*network.Peer
    blockHeight int
    block *common.Block
}

func GetState() *State {
    once.Do(func() {
        singleton = new(State)
    })
    return singleton
}

func (s *State) ChangeRole(role int) {
    lock.Lock()
    s.role = role
    lock.Unlock()
}

func (s *State) GetRole() int {
    return s.role
}

func (s *State) ChangeMiningState(state int) {
    lock.Lock()
    s.miningState = state
    lock.Unlock()
}

func (s *State) GetMiningState() int {
    return s.miningState
}

func (s *State) GetMiners() []*Miner {
    return s.miners
}

func (s *State) ContainsMiner(pubKey []byte) bool {
    for _, v:= range s.miners {
        if util.SliceEqual(v.PubKey, pubKey) {
            return true
        }
    }
    return false
}

func (s *State) GetMiner(pubKey []byte) *Miner {
    for _, v:= range s.miners {
        if util.SliceEqual(v.PubKey, pubKey) {
            return v
        }
    }
    return nil
}

func (s *State) GetPeers() []*network.Peer {
    return s.peers
}

func (s *State) GetBlockHeight() int {
    return s.blockHeight
}

func (s *State) CheckState(role int, miningState int) bool {
    return s.role == role && s.miningState == miningState
}

func (s *State) MoveToState(miningState int) {
    s.miningState = miningState
}

func (s *State) GetLeader() *Miner {
    return nil
}

func (s *State) GetBlock() *common.Block {
    return s.block
}

func (s *State) SetBlock(block *common.Block) {
    s.block = block
}
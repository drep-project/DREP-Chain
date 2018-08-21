package storage

import (
    "sync"
    "BlockChainTest/util"
    "BlockChainTest/network"
)

const (
    Lead        int = 0
    MINER       int = 1
    NonMiner    int = 2
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
    Address string
}

type State struct {
    role int
    miningState int
    miners []*Miner
    peers []*network.Peer
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

func (s *State) GetPeers() []*network.Peer {
    return s.peers
}
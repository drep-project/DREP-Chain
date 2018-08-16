package storage

import "sync"

const (
    Lead        int = 0
    Miner       int = 1
    NonMiner    int = 2
)

const (

)

var (
    once sync.Once
    singleton *State
    lock sync.Locker
)

type State struct {
    role int
    miningState int
    miners [][]byte

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

func (s *State) GetMiners() [][]byte {
    return s.miners
}
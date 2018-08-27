package node

import (
    "BlockChainTest/network"
    "BlockChainTest/bean"
)

const (
    LEADER     = 0
    MINER      = 1
    NON_MINER  = 2
)

type Miner struct {
    PubKey *bean.Point
    Address bean.Address
    Peer *network.Peer
}


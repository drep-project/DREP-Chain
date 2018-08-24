package node

import (
    "BlockChainTest/network"
    "BlockChainTest/common"
    "BlockChainTest/bean"
)

const (
    LEADER     = 0
    MINER      = 1
    NON_MINER  = 2
)

type Miner struct {
    PubKey *bean.Point
    Address common.Address
    Peer *network.Peer
}


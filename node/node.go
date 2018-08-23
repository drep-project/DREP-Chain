package node

import (
    "BlockChainTest/network"
    "BlockChainTest/common"
)

const (
    LEADER     = 0
    MINER      = 1
    NON_MINER  = 2
)

type Miner struct {
    PubKey []byte
    Address common.Address
    Peer *network.Peer
}


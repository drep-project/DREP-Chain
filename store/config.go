package store

import (
    "math/big"
    "BlockChainTest/network"
)

var (
    BlockGasLimit  = big.NewInt(50)
    GasPrice = big.NewInt(5)
    TransferGas = big.NewInt(10)
    MinerGas = big.NewInt(10)
    TransferType int32 = 0
    MinerType int32 = 1
    CreateContractType int32 = 2
    CallContractType int32 = 3
    // TODO
    Admin *network.Peer
    Version int32 = 1
)

var IsStart bool

const LocalTest = false

func init() {
    if LocalTest {
        Admin = &network.Peer{IP: network.IP("127.0.0.1"), Port: 55555}
    } else {
        Admin = &network.Peer{IP: network.IP("192.168.3.231"), Port: 55555}
    }
}


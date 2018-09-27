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
    // TODO
    Admin = &network.Peer{IP: network.IP("192.168.3.147"), Port: 55555}
)

const IsStart = true
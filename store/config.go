package store

import (
    "math/big"
    "BlockChainTest/network"
)

var (
    BlockGasLimit  = big.NewInt(50)
    GasPrice = big.NewInt(5)
    TransferGas = big.NewInt(10)
    // TODO
    Admin = &network.Peer{IP: network.IP("192.168"), Port: 55555}
)

const IsAdmin = true
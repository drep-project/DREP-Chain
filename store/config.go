package store

import (
    "math/big"
    "BlockChainTest/bean"
)

var (
    BlockGasLimit  = big.NewInt(5000000000)
    GasPrice = big.NewInt(5)
    TransferGas = big.NewInt(10)
    MinerGas = big.NewInt(10)
    CreateContractGas = big.NewInt(1000)
    CallContractGas = big.NewInt(100000)
    TransferType int32 = 0
    MinerType int32 = 1
    CreateContractType int32 = 2
    CallContractType int32 = 3
    // TODO
    Admin *bean.Peer
    Version int32 = 1
)

var IsStart bool

const LocalTest = false

func init() {
    if LocalTest {
        Admin = &bean.Peer{IP: bean.IP("127.0.0.1"), Port: 55555}
    } else {
        Admin = &bean.Peer{IP: bean.IP("192.168.3.231"), Port: 55555}
    }
}


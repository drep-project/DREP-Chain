package store

import (
    "math/big"
    "BlockChainTest/bean"
)

var (
    GWei                     = new(big.Int).SetInt64(1000000000)
    BlockGasLimit            = big.NewInt(5000000000)
    DefaultGasPrice          = big.NewInt(10)
    TransferGas              = big.NewInt(10)
    MinerGas                 = big.NewInt(10)
    CreateContractGas        = big.NewInt(10)
    CallContractGas          = big.NewInt(10)
    TransferType       int32 = 0
    MinerType          int32 = 1
    CreateContractType int32 = 2
    CallContractType   int32 = 3
    Version            int32 = 1
    // TODO
    Admin *bean.Peer
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


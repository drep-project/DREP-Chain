package store

import (
    "math/big"
    "BlockChainTest/bean"
)

var (
    BlockGasLimit            = big.NewInt(5000000000)
    DefaultGasPrice          *big.Int
    TransferGas              = big.NewInt(20000)
    MinerGas                 = big.NewInt(20000)
    CreateContractGas        = big.NewInt(1000000)
    CallContractGas          = big.NewInt(10000000)
    TransferType       int32 = 0
    MinerType          int32 = 1
    CreateContractType int32 = 2
    CallContractType   int32 = 3
    Version            int32 = 1
    // TODO
    Admin *bean.Peer
)

const LocalTest = false

func init() {
    DefaultGasPrice, _ = new(big.Int).SetString("20000000000", 10)
    if LocalTest {
        Admin = &bean.Peer{IP: bean.IP("127.0.0.1"), Port: 55555}
    } else {
        Admin = &bean.Peer{IP: bean.IP("192.168.3.231"), Port: 55555}
    }
}


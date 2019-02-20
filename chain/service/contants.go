package service

import (
    "math/big"
)

var (
    BlockGasLimit            = big.NewInt(5000000000000000000)
    DefaultGasPrice          *big.Int
    TransferGas              = big.NewInt(20)
    MinerGas                 = big.NewInt(20000)
    CreateContractGas        = big.NewInt(1000000)
    CallContractGas          = big.NewInt(10000000)
    CrossChainGas = big.NewInt(10000000)
    TransferType       int32 = 0
    MinerType          int32 = 1
    CreateContractType int32 = 2
    CallContractType   int32 = 3
    CrossChainType     int32 = 4
    BlockPrizeType     int32 = 5
    Version            int32 = 1
)

const LocalTest = false

func init() {
    DefaultGasPrice, _ = new(big.Int).SetString("20000000000", 10)
}


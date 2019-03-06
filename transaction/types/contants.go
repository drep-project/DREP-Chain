package types

import "math/big"

type TxType  int
const(
	TransferType TxType = iota
	MinerType
	CreateContractType
	CallContractType
	CrossChainType
)


var (
	DefaultGasPrice    *big.Int
	TransferGas              = big.NewInt(20)
	MinerGas                 = big.NewInt(20000)
	CreateContractGas        = big.NewInt(1000000)
	CallContractGas          = big.NewInt(10000000)
	CrossChainGas            = big.NewInt(10000000)
)

const LocalTest = false

func init() {
	DefaultGasPrice, _ = new(big.Int).SetString("20000000000", 10)
}

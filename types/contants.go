package types

import "math/big"

type TxType uint64

const (
	TransferType TxType = iota
	_
	CreateContractType
	CallContractType
	_
	SetAliasType //给地址设置昵称
)

var (
	TransferGas       = big.NewInt(30000)
	MinerGas          = big.NewInt(20000)
	CreateContractGas = big.NewInt(1000000)
	CallContractGas   = big.NewInt(10000000)
	CrossChainGas     = big.NewInt(10000000)
	SeAliasGas        = big.NewInt(10000000)
)

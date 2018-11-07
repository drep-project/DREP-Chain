package bean

import "math/big"

type Account struct {
	Addr                 []byte
	Nonce                int64
	Balance             *big.Int
	IsContract           bool
	ByteCode             []byte
	CodeHash             []byte
}

type Log struct {
	CallerAddr           []byte
	ContractAddr         []byte
	TxHash               []byte
	Topics               [][]byte
	Data                 []byte
}
package params

const (
	Wei     = 1
	GWei    = 1e9
	Coin    = 1e18
	Rewards = 1000000000000000000 //每出一个块，系统奖励的币数目

	AliasGas uint64 = 42000 // gas Use when alias a address
	TxGas                 uint64 = 21000 // Per transaction not creating a contract. NOTE: Not payable on data of calls between transactions.
	TxDataZeroGas         uint64 = 4     // Per byte of data attached to a transaction that equals zero. NOTE: Not payable on data of calls between transactions.
	TxDataNonZeroGas uint64 = 68    // Per byte of data attached to a transaction that is not equal to zero. NOTE: Not payable on data of calls between transactions.
	MinGasLimit          uint64 = 180000000 // Minimum the gas limit may ever be.
	GenesisGasLimit      uint64 = 180000000 // Gas limit of the Genesis block.
	//MIN_GAS_IN_BLOCK uint64 = 60000000 / 2
	MaxGasLimit uint64 = 360000000 // tps 3000 transfer  60000gas per transfer tx
	MaxSupply   uint64 = 10000000000
)

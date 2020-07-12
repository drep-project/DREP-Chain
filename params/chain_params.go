package params

const (
	Wei                          = 1
	GWei                         = 1e9
	Coin                         = 1e18
	Rewards                      = 100 //每出一个块，系统奖励的币数目，单位1drep
	BlockCountOfEveryYear uint64 = 2102400

	AliasGas uint64 = 68 // gas Use when alias a address

	//GasLimitBoundDivisor uint64 = 64       // The bound divisor of the gas limit, used in update calculations.
	MinGasLimit     uint64 = 18000000 // Minimum the gas limit may ever be.
	GenesisGasLimit uint64 = 18000000 // Gas limit of the Genesis block.
	MaxGasLimit     uint64 = 70000000 // tps 3000 transfer  60000gas per transfer tx
	//MIN_GAS_IN_BLOCK uint64 = 60000000 / 2

	//MaximumExtraDataSize  uint64 = 32    // Maximum size extra data may be after Genesis.
	//ExpByteGas            uint64 = 10    // Times ceil(log256(exponent)) for the EXP instruction.
	//SloadGas              uint64 = 50    // Mullied by the number of 32-byte words that are copied (round up) for any *COPY operation and added.
	CallValueTransferGas  uint64 = 9000  // Paid for CALL when the value transfer is non-zero.
	CallNewAccountGas     uint64 = 25000 // Paid for CALL when the destination address didn't exist prior.
	TxGas                 uint64 = 21000 // Per transaction not creating a contract. NOTE: Not payable on data of calls between transactions.
	TxGasContractCreation uint64 = 53000 // Per transaction that creates a contract. NOTE: Not payable on data of calls between transactions.
	TxDataZeroGas         uint64 = 4     // Per byte of data attached to a transaction that equals zero. NOTE: Not payable on data of calls between transactions.
	QuadCoeffDiv          uint64 = 512   // Divisor for the quadratic particle of the memory cost equation.
	LogDataGas            uint64 = 8     // Per byte in a LOG* operation's data.
	CallStipend           uint64 = 2300  // Free gas given at beginning of call.

	Sha3Gas     uint64 = 30 // Once per SHA3 operation.
	Sha3WordGas uint64 = 6  // Once per word of the SHA3 operation's data.

	SstoreSetGas    uint64 = 20000 // Once per SLOAD operation.
	SstoreResetGas  uint64 = 5000  // Once per SSTORE operation if the zeroness changes from zero.
	SstoreClearGas  uint64 = 5000  // Once per SSTORE operation if the zeroness doesn't change.
	SstoreRefundGas uint64 = 15000 // Once per SSTORE operation if the zeroness changes to zero.

	//NetSstoreNoopGas  uint64 = 200   // Once per SSTORE operation if the value doesn't change.
	//NetSstoreInitGas  uint64 = 20000 // Once per SSTORE operation from clean zero.
	//NetSstoreCleanGas uint64 = 5000  // Once per SSTORE operation from clean non-zero.
	//NetSstoreDirtyGas uint64 = 200   // Once per SSTORE operation from dirty.
	//
	//NetSstoreClearRefund      uint64 = 15000 // Once per SSTORE operation for clearing an originally existing storage slot
	//NetSstoreResetRefund      uint64 = 4800  // Once per SSTORE operation for resetting to the original non-zero value
	//NetSstoreResetClearRefund uint64 = 19800 // Once per SSTORE operation for resetting to the original zero value

	JumpdestGas uint64 = 1 // Refunded gas, once per SSTORE operation if the zeroness changes to zero.
	//EpochDuration    uint64 = 30000 // Duration between proof-of-work epochs.
	//CallGas          uint64 = 40    // Once per CALL operation & message call transaction.
	CreateDataGas   uint64 = 200  //
	CallCreateDepth uint64 = 1024 // Maximum depth of call/create stack.
	//ExpGas           uint64 = 10    // Once per EXP instruction
	LogGas     uint64 = 375  // Per LOG* operation.
	CopyGas    uint64 = 3    //
	StackLimit uint64 = 1024 // Maximum size of VM stack allowed.
	//TierStepGas      uint64 = 0     // Once per operation, for a selection of them.
	LogTopicGas      uint64 = 375   // Multiplied by the * of the LOG*, per LOG transaction. e.g. LOG0 incurs 0 * c_txLogTopicGas, LOG4 incurs 4 * c_txLogTopicGas.
	CreateGas        uint64 = 32000 // Once per CREATE operation & contract-creation transaction.
	Create2Gas       uint64 = 32000 // Once per CREATE2 operation
	SuicideRefundGas uint64 = 24000 // Refunded following a suicide operation.
	MemoryGas        uint64 = 3     // Times the address of the (highest referenced byte in memory + 1). NOTE: referencing happens on read, write and in instructions such as RETURN and CALL.
	TxDataNonZeroGas uint64 = 68    // Per byte of data attached to a transaction that is not equal to zero. NOTE: Not payable on data of calls between transactions.

	MaxCodeSize = 24576 // Maximum bytecode to permit for a contract

	// Precompiled contract gas prices

	EcrecoverGas            uint64 = 3000   // Elliptic curve sender recovery gas price
	Sha256BaseGas           uint64 = 60     // Base price for a SHA256 operation
	Sha256PerWordGas        uint64 = 12     // Per-word price for a SHA256 operation
	Ripemd160BaseGas        uint64 = 600    // Base price for a RIPEMD160 operation
	Ripemd160PerWordGas     uint64 = 120    // Per-word price for a RIPEMD160 operation
	IdentityBaseGas         uint64 = 15     // Base price for a data copy operation
	IdentityPerWordGas      uint64 = 3      // Per-work price for a data copy operation
	ModExpQuadCoeffDiv      uint64 = 20     // Divisor for the quadratic particle of the big int modular exponentiation
	Bn256AddGas             uint64 = 500    // Gas needed for an elliptic curve addition
	Bn256ScalarMulGas       uint64 = 40000  // Gas needed for an elliptic curve scalar multiplication
	Bn256PairingBaseGas     uint64 = 100000 // Base price for an elliptic curve pairing check
	Bn256PairingPerPointGas uint64 = 80000  // Per-point price for an elliptic curve pairing check

	RootChain                 uint32 = 0
	RemotePortMainnet         uint16 = 10087
	GenesisProducerNumMainnet        = 21

	RemotePortTestnet         uint16 = 44445
	GenesisProducerNumTestnet        = 3

	BlockInterval  int16  = 15
	ChangeInterval uint64 = 100
)

var (
	BootStrapNodeMainnet = []string{"enode://f57881c48aaccf97485c2b65b421bfeda22cc3b427c44be7607b122fc1688abb@172.104.123.143:10086",
		"enode://9d25d161ae4b676e2df55accca93c3137df3166326d04420ffbdf66e887bd494@172.104.116.219:10086",
		"enode://bc7ca1b57175f2d5c85da73d367408529468a034b97d083aaecf88196090e245@172.105.103.59:10086",
		"enode://0ebd0422ca32d70292be128342f9e5ca32ab3cef28dc32cc332169e578e7b4f5@109.74.203.199:10086",
	}

	BootStrapNodeTestnet = []string{"enode://548c58daf6dc65d463c155027fce3a909d555683543d1dca34cff1d68868c54f@39.100.111.74:44444",
		"enode://385c49f05a235115515d5581485be6cd66bbcaf2dbace93d641b5e4c87c20255@39.98.39.224:44444",
		"enode://9296c4f6e4ceaaea24d0416f49bf7624e920d1f71f7a51877a5d0ed156e35ac5@39.99.44.60:44444",
	}
)

type NetType string

const (
	TestnetType NetType = "testnet"
	MainnetType NetType = "mainnet"
	SolonetType NetType = "solonet" //develop net
)

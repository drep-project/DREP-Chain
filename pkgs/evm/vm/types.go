package vm

import (
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/types"
	"math/big"
)

type VMConfig struct {
	// Debug enabled debugging Interpreter options
	LogConfig *LogConfig `json:"logconfig"`
	// Tracer is the op code logger
	//	Tracer Tracer
	// NoRecursion disabled Interpreter call, callcode,
	// delegate call and create.
	NoRecursion bool `json:"noRecursion"`
	// Miner recording of SHA3/keccak preimages
	EnablePreimageRecording bool `json:"enablePreimageRecording"`
	// JumpTable contains the EVM instruction table. This
	// may be left uninitialised and will be set to the default
	// table.
	//JumpTable [256]vm.operation

	// Type of the EWASM interpreter
	EWASMInterpreter string `json:"ewasmInterpreter"`
	// Type of the EVM interpreter
	EVMInterpreter string `json:"evmInterpreter"`
}

// LogConfig are the configuration options for structured logger the EVM
type LogConfig struct {
	DisableMemory  bool // disable memory capture
	DisableStack   bool // disable stack capture
	DisableStorage bool // disable storage capture
	Debug          bool // print output during capture end
	Limit          int  // maximum length of output, but zero means unlimited
}

type Message struct {
	From      crypto.CommonAddress
	To        crypto.CommonAddress
	ChainId   types.ChainIdType
	DestChain types.ChainIdType
	Gas       uint64
	Value     *big.Int
	Nonce     uint64
	Input     []byte
	ReadOnly  bool
}

package vm

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/crypto"
	"math/big"
)

type VMConfig struct {
	// Debug enabled debugging Interpreter options
	Debug bool							  `json:"debug"`
	// Tracer is the op code logger
//	Tracer Tracer
	// NoRecursion disabled Interpreter call, callcode,
	// delegate call and create.
	NoRecursion bool					  `json:"noRecursion"`
	// Enable recording of SHA3/keccak preimages
	EnablePreimageRecording bool		  `json:"enablePreimageRecording"`
	// JumpTable contains the EVM instruction table. This
	// may be left uninitialised and will be set to the default
	// table.
	//JumpTable [256]vm.operation

	// Type of the EWASM interpreter
	EWASMInterpreter string				  `json:"ewasmInterpreter"`
	// Type of the EVM interpreter
	EVMInterpreter string				  `json:"evmInterpreter"`
}


type Message struct {
	From      crypto.CommonAddress
	To        crypto.CommonAddress
	ChainId   app.ChainIdType
	DestChain app.ChainIdType
	Gas       uint64
	Value     *big.Int
	Nonce     uint64
	Input     []byte
	ReadOnly  bool
}
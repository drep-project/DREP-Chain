// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/params"
	"github.com/drep-project/DREP-Chain/types"
	"math/big"
	"sync/atomic"
	"time"
)

// emptyCodeHash is used by create to ensure deployment is disallowed to already
// deployed contract addresses (relevant after the account abstraction).
var emptyCodeHash = crypto.Keccak256Hash(nil)

type (
	// CanTransferFunc is the signature of a transfer guard function
	CanTransferFunc func(VMState, crypto.CommonAddress, *big.Int) bool
	// TransferFunc is the signature of a transfer function
	TransferFunc func(VMState, crypto.CommonAddress, crypto.CommonAddress, *big.Int)
	// GetHashFunc returns the nth block hash in the blockchain
	// and is used by the BLOCKHASH EVM op code.
	GetHashFunc func(uint64) crypto.Hash
)

// run runs the given contract and takes care of running precompiles with a fallback to the byte code interpreter.
func run(evm *EVM, contract *Contract, input []byte, readOnly bool) ([]byte, error) {
	if !contract.ContractAddr.IsEmpty() {
		precompiles := PrecompiledContracts
		//原生合约
		if p := precompiles[contract.ContractAddr]; p != nil {
			return RunPrecompiledContract(p, input, contract)
		}
	}

	interpreter := evm.Interpreter()
	if interpreter.CanRun(contract.ByteCode) {
		return interpreter.Run(contract, input, readOnly)
	}

	/*
		if interpreter.CanRun(contract.ByteCode) {
			if evm.interpreter != interpreter {
				// Ensure that the interpreter pointer is set back
				// to its current value upon return.
				defer func(i Interpreter) {
					evm.interpreter = i
				}(evm.interpreter)
				evm.interpreter = interpreter
			}
			return interpreter.Run(contract, input, readOnly)
		}
	*/
	return nil, ErrNoCompatibleInterpreter
}

// Context provides the EVM with auxiliary information. Once provided
// it shouldn't be modified.
type Context struct {
	// CanTransfer returns whether the account contains
	// sufficient ether to transfer the value
	CanTransfer CanTransferFunc
	// Transfer transfers ether from one account to the other
	Transfer TransferFunc
	// GetHash returns the hash corresponding to n
	GetHash GetHashFunc

	// Message information
	Origin   crypto.CommonAddress // Provides information for ORIGIN
	GasPrice *big.Int             // Provides information for GASPRICE

	// Block information
	GasLimit    uint64   // Provides information for GASLIMIT
	BlockNumber *big.Int // Provides information for NUMBER
	Time        *big.Int // Provides information for TIME
	TxHash      *crypto.Hash
}

// EVM is the Ethereum Virtual Machine base object and provides
// the necessary tools to run a contract on the given state with
// the provided context. It should be noted that any error
// generated through any of the calls should be considered a
// revert-state-and-consume-all-gas operation, no checks on
// specific errors should ever be performed. The interpreter makes
// sure that any errors generated are to be considered faulty code.
//
// The EVM should never be reused and is not thread safe.
type EVM struct {
	// Context provides auxiliary blockchain related information
	Context
	// StateDB gives access to the underlying state
	State VMState
	// Depth is the current call stack
	depth int

	// virtual machine configuration options used to initialise the
	// evm.
	vmConfig *VMConfig
	// global (to this context) ethereum virtual machine
	// used throughout the execution of the tx.
	interpreter *EVMInterpreter
	// abort is used to abort the EVM calling operations
	// NOTE: must be set atomically
	abort int32
	// callGasTemp holds the gas available for the current call. This is needed because the
	// available gas is calculated in gasCall* according to the 63/64 rule and later
	// applied in opCall*.
	CallGasTemp uint64

	ChainId types.ChainIdType
}

// NewEVM returns a new EVM. The returned EVM is not thread safe and should
// only ever be used *once*.
func NewEVM(ctx Context, statedb VMState, vmConfig *VMConfig) *EVM {
	evm := &EVM{
		Context:  ctx,
		State:    statedb,
		vmConfig: vmConfig,
	}

	evm.interpreter = NewEVMInterpreter(evm)

	return evm
}

// Cancel cancels any running EVM operation. This may be called concurrently and
// it's safe to be called multiple times.
func (evm *EVM) Cancel() {
	atomic.StoreInt32(&evm.abort, 1)
}

// Interpreter returns the current interpreter
func (evm *EVM) Interpreter() *EVMInterpreter {
	return evm.interpreter
}

// Call executes the contract associated with the addr with the given input as
// parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
func (evm *EVM) Call(caller crypto.CommonAddress, addr crypto.CommonAddress, chainId types.ChainIdType, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, gas, nil
	}

	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}
	// Fail if we're trying to transfer more than the available balance
	if !evm.Context.CanTransfer(evm.State, caller, value) {
		return nil, gas, ErrInsufficientBalance
	}

	var (
		to = addr
	)
	if !evm.State.Exist(addr) {
		precompiles := PrecompiledContracts
		if precompiles[addr] == nil && value.Sign() == 0 {
			// Calling a non existing account, don't do anything, but ping the tracer
			//if evm.vmConfig.Debug && evm.depth == 0 {
			//	evm.vmConfig.Tracer.CaptureStart(caller.Address(), addr, false, input, gas, value)
			//	evm.vmConfig.Tracer.CaptureEnd(ret, 0, 0, nil)
			//}
			return nil, gas, nil
		}

		//evm.State.CreateAccount(addr)
		return nil, 0, ErrNoAccount
	}
	evm.Transfer(evm.State, caller, to, value)
	// Initialise a new contract and set the code that is to be used by the EVM.
	// The contract is a scoped environment for this execution context only.
	contract := NewContract(caller, evm.TxHash, chainId, gas, value, nil)
	contract.SetCode(addr, evm.State.GetByteCode(&addr))
	// Even if the account has no code, we need to continue because it might be a precompile
	start := time.Now()

	// Capture the tracer start/end events in debug mode

	if evm.vmConfig.LogConfig.Debug && evm.depth == 0 {
		evm.interpreter.Tracer.CaptureStart(caller, addr, false, input, gas, value)

		defer func() { // Lazy evaluation of the parameters
			evm.interpreter.Tracer.CaptureEnd(ret, gas-contract.Gas, time.Since(start), err)
		}()
	}

	ret, err = run(evm, contract, input, false)

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in homestead this also counts for code storage gas errors.
	if err != nil {
		if err != errExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}
	return ret, contract.Gas, err
}

// CallCode executes the contract associated with the addr with the given input
// as parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
//
// CallCode differs from Call in the sense that it executes the given address'
// code with the caller as context.
func (evm *EVM) CallCode(caller crypto.CommonAddress, addr crypto.CommonAddress, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, gas, nil
	}

	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}
	// Fail if we're trying to transfer more than the available balance
	if !evm.CanTransfer(evm.State, caller, value) {
		return nil, gas, ErrInsufficientBalance
	}

	// initialise a new contract and set the code that is to be used by the
	// EVM. The contract is a scoped environment for this execution context
	// only.
	contract := NewContract(caller, evm.TxHash, evm.ChainId, gas, value, nil)
	contract.SetCode(addr, evm.State.GetByteCode(&addr))

	ret, err = run(evm, contract, input, false)
	if err != nil {
		if err != errExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}
	return ret, contract.Gas, err
}

func (evm *EVM) DelegateCall(con *Contract, contractAddr crypto.CommonAddress, input []byte, gas uint64) (ret []byte, leftGas uint64, err error) {
	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, gas, nil
	}
	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}

	callerAddr := con.CallerAddr
	chainId := con.ChainId
	jumpdests := con.Jumpdests

	byteCode := evm.State.GetByteCode(&contractAddr)
	if byteCode == nil {
		return nil, gas, ErrCodeNotExists
	}

	contract := NewContract(callerAddr, evm.TxHash, chainId, gas, new(big.Int), jumpdests)
	contract.SetCode(contractAddr, byteCode)

	ret, err = run(evm, contract, input, false)
	if err != nil {
		//evm.State.dt.Discard()
		if err != errExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	} else {
		//evm.State.dt.TrieDb()
	}

	return ret, con.Gas, err
}

// StaticCall executes the contract associated with the addr with the given input
// as parameters while disallowing any modifications to the state during the call.
// Opcodes that attempt to perform such modifications will result in exceptions
// instead of performing the modifications.
func (evm *EVM) StaticCall(caller crypto.CommonAddress, addr crypto.CommonAddress, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, gas, nil
	}
	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}

	byteCode := evm.State.GetByteCode(&caller)
	if byteCode == nil {
		return nil, gas, ErrCodeNotExists
	}

	contract := NewContract(caller, evm.TxHash, evm.ChainId, gas, new(big.Int), nil)
	contract.SetCode(addr, byteCode)

	// We do an AddBalance of zero here, just in order to trigger a touch.
	// This doesn't matter on Mainnet, where all empties are gone at the time of Byzantium,
	// but is the correct thing to do and matters on other networks, in tests, and potential
	// future scenarios
	evm.State.AddBalance(&addr, bigZero)

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in Homestead this also counts for code storage gas errors.
	ret, err = run(evm, contract, input, true)
	if err != nil {
		if err != errExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}
	return ret, contract.Gas, err
}

type codeAndHash struct {
	code []byte
	hash crypto.Hash
}

func (c *codeAndHash) Hash() crypto.Hash {
	if c.hash == (crypto.Hash{}) {
		c.hash = crypto.Keccak256Hash(c.code)
	}
	return c.hash
}

// create creates a new contract using code as deployment code.
func (evm *EVM) CreateContractCode(caller crypto.CommonAddress, codeAndHash *codeAndHash, gas uint64, value *big.Int, address crypto.CommonAddress) ([]byte, crypto.CommonAddress, uint64, error) {
	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	if evm.depth > int(params.CallCreateDepth) {
		return nil, crypto.CommonAddress{}, gas, ErrDepth
	}
	if !evm.CanTransfer(evm.State, caller, value) {
		return nil, crypto.CommonAddress{}, gas, ErrInsufficientBalance
	}
	nonce := evm.State.GetNonce(&caller)
	evm.State.SetNonce(&caller, nonce+1)
	// Ensure there's no existing contract already at the designated address
	contractHash := evm.State.GetCodeHash(address)
	if evm.State.GetNonce(&address) != 0 || (contractHash != (crypto.Hash{}) && contractHash != emptyCodeHash) {
		return nil, crypto.CommonAddress{}, 0, ErrContractAddressCollision
	}
	// Create a new account on the state
	account, err := evm.State.CreateContractAccount(address, codeAndHash.code)
	evm.Transfer(evm.State, caller, address, value)

	// initialise a new contract and set the code that is to be used by the
	// EVM. The contract is a scoped environment for this execution context
	// only.
	contractAddr := account.Address
	contract := NewContract(caller, evm.TxHash, evm.ChainId, gas, value, nil)
	contract.SetCode(*contractAddr, codeAndHash.code)

	if evm.vmConfig.NoRecursion && evm.depth > 0 {
		return nil, address, gas, nil
	}

	if evm.vmConfig.LogConfig.Debug && evm.depth == 0 {
		//evm.vmConfig.Tracer.CaptureStart(caller, address, true, codeAndHash.code, gas, value)
	}

	ret, err := run(evm, contract, nil, false)

	// check whether the max code size has been exceeded
	maxCodeSizeExceeded := len(ret) > params.MaxCodeSize
	// if the contract creation ran successfully and no errors were returned
	// calculate the gas required to store the code. If the code could not
	// be stored due to not enough gas set an error and let it be handled
	// by the error checking condition below.
	if err == nil && !maxCodeSizeExceeded {
		createDataGas := uint64(len(ret)) * params.CreateDataGas
		if contract.UseGas(createDataGas) {
			evm.State.SetByteCode(contractAddr, ret)
		} else {
			err = ErrCodeStoreOutOfGas
		}
	}

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in homestead this also counts for code storage gas errors.
	if maxCodeSizeExceeded || (err != nil || err != ErrCodeStoreOutOfGas) {
		//evm.State.RevertToSnapshot(snapshot)
		if err != errExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}
	// Assign err if contract code size exceeds the max while the err is still empty.
	if maxCodeSizeExceeded && err == nil {
		err = errMaxCodeSizeExceeded
	}
	if evm.vmConfig.LogConfig.Debug && evm.depth == 0 {
		//evm.vmConfig.Tracer.CaptureEnd(ret, gas-contract.Gas, time.Since(start), err)
	}
	return ret, address, contract.Gas, err
}

// Create creates a new contract using code as deployment code.
func (evm *EVM) Create(caller crypto.CommonAddress, code []byte, gas uint64, value *big.Int) (ret []byte, contractAddr crypto.CommonAddress, leftOverGas uint64, err error) {
	contractAddr = crypto.CreateAddress(caller, evm.State.GetNonce(&caller))

	return evm.CreateContractCode(caller, &codeAndHash{code: code}, gas, value, contractAddr)
}

// Create2 creates a new contract using code as deployment code.
//
// The different between Create2 with Create is Create2 uses sha3(0xff ++ msg.sender ++ salt ++ sha3(init_code))[12:]
// instead of the usual sender-and-nonce-hash as the address where the contract is initialized at.
//func (evm *EVM) Create2(caller crypto.CommonAddress, code []byte, gas uint64, endowment *big.Int, salt *big.Int) (ret []byte, contractAddr crypto.CommonAddress, leftOverGas uint64, err error) {
//	codeAndHash := &codeAndHash{code: code}
//	contractAddr = crypto.CreateAddress2(caller, crypto.BigToHash(salt), codeAndHash.Hash().Bytes())
//	return evm.CreateContractCode(caller, codeAndHash, gas, endowment, contractAddr)
//}

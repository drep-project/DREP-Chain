package vm

import (
	"math/big"
	"errors"
	"fmt"
	"BlockChainTest/bean"
)

type EVM struct {
	State *State
	Interpreter *EVMInterpreter
	CallGasTemp uint64
	GasLimit uint64
	Origin bean.CommonAddress
	GasPrice *big.Int
	CoinBase bean.CommonAddress
	Time *big.Int
	Abort int32
}

var (
	ErrNoCompatibleInterpreter  = errors.New("no compatible interpreter")
)

func NewEVM() *EVM {
	evm := &EVM{}
	evm.State = GetState()
	evm.Interpreter = NewEVMInterpreter(evm)
	return evm
}

func (evm *EVM) CreateContractCode(callerAddr bean.CommonAddress, byteCode bean.ByteCode, gas uint64, value *big.Int) ([]byte, bean.CommonAddress, error) {
	if !evm.CanTransfer(callerAddr, value) {
		return nil, bean.CommonAddress{}, ErrInsufficientBalance
	}

	contractAddr := bean.CodeAddr(byteCode)
	_, err := evm.State.GetByteCode(contractAddr)
	if err == nil {
		return nil, bean.CommonAddress{}, ErrCodeAlreadyExists
	}

	nonce, err := evm.State.GetNonce(callerAddr)
	if err != nil {
		return nil, bean.CommonAddress{}, err
	}
	evm.State.SetNonce(callerAddr, nonce + 1)
	evm.Transfer(callerAddr, contractAddr, value)
	evm.State.CreateContractAccount(contractAddr, byteCode)

	return nil, contractAddr, err
}

func (evm *EVM) CallContractCode(callerAddr, contractAddr bean.CommonAddress, input []byte, gas uint64, value *big.Int) (ret []byte, returnGas uint64, err error) {
	if !evm.CanTransfer(callerAddr, value) {
		return nil, gas, ErrInsufficientBalance
	}

	byteCode, err := evm.State.GetByteCode(contractAddr)
	if byteCode == nil {
		return nil, gas, ErrCodeNotExists
	}
	evm.Transfer(callerAddr, contractAddr, value)

	codeHash := bean.CodeHash(byteCode)
	contract := NewContract(callerAddr, gas, value, nil)
	contract.SetCode(contractAddr, codeHash, byteCode)

	ret, err = run(evm, contract, input)
	return ret, contract.Gas, err
}

func (evm *EVM) StaticCall(callerAddr, contractAddr bean.CommonAddress, input []byte, gas uint64) (ret []byte, returnGas uint64, err error) {
	byteCode, err := evm.State.GetByteCode(contractAddr)
	if byteCode == nil {
		return nil, gas, ErrCodeNotExists
	}
	evm.Transfer(callerAddr, contractAddr, new(big.Int))

	codeHash := bean.CodeHash(byteCode)
	contract := NewContract(callerAddr, gas, new(big.Int), nil)
	contract.SetCode(contractAddr, codeHash, byteCode)

	ret, err = run(evm, contract, input)
	return ret, contract.Gas, err
}

func (evm *EVM) DelegateCall(con *Contract, contractAddr bean.CommonAddress, input []byte, gas uint64) (ret []byte, leftGas uint64, err error) {
	callerAddr := con.CallerAddr
	jumpdests := con.Jumpdests

	byteCode, err := evm.State.GetByteCode(contractAddr)
	if byteCode == nil {
		return nil, gas, ErrCodeNotExists
	}

	codeHash := bean.CodeHash(byteCode)
	contract := NewContract(callerAddr, gas, new(big.Int), jumpdests)
	contract.SetCode(contractAddr, codeHash, byteCode)

	ret, err = run(evm, contract, input)
	return ret, con.Gas, err
}


func run(evm *EVM, contract *Contract, input []byte) ([]byte, error) {
	if !contract.ContractAddr.IsEmpty() {
		precompiles := PrecompiledContracts
		if p := precompiles[contract.ContractAddr]; p != nil {
			return RunPrecompiledContract(p, input, contract)
		}
	}
	interpreter := evm.Interpreter
	fmt.Println()
	fmt.Println("interpreter: ", interpreter)
	fmt.Println()
	if interpreter.CanRun(contract.ByteCode) {
		return interpreter.Run(contract, input)
	}
	return nil, ErrNoCompatibleInterpreter
}

func (evm *EVM) CanTransfer(addr bean.CommonAddress, amount *big.Int) bool {
	balance, err := evm.State.GetBalance(addr)
	fmt.Println("balance: ", balance)
	fmt.Println("balance error: ", err)
	if err != nil {
		return false
	}
	return balance.Cmp(amount) >= 0
}

func (evm *EVM) Transfer(from, to bean.CommonAddress, amount *big.Int) error {
	err := evm.State.SubBalance(from, amount)
	if err != nil {
		return err
	}
	err = evm.State.AddBalance(to, amount)
	if err != nil {
		evm.State.AddBalance(from, amount)
		return err
	}
	return nil
}


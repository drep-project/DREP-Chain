package vm

import (
	"math/big"
	"errors"
	"BlockChainTest/bean"
	"BlockChainTest/accounts"
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

func (evm *EVM) CreateContractCode(callerAddr accounts.CommonAddress, chainId int64, byteCode accounts.ByteCode, gas uint64, value *big.Int) ([]byte, accounts.CommonAddress, error) {
	if !evm.CanTransfer(callerAddr, chainId, value) {
		return nil, accounts.CommonAddress{}, ErrInsufficientBalance
	}

	nonce := evm.State.GetNonce(callerAddr, chainId)
	account, err := evm.State.CreateContractAccount(callerAddr, chainId, nonce, byteCode)
	if err != nil {
		return nil, accounts.CommonAddress{}, err
	}

	contractAddr := account.Address
	evm.State.SetNonce(callerAddr, chainId, nonce + 1)
	evm.Transfer(callerAddr, contractAddr, chainId, value)

	return nil, contractAddr, nil
}

func (evm *EVM) CallContractCode(callerAddr, contractAddr accounts.CommonAddress, chainId int64, input []byte, gas uint64, value *big.Int) (ret []byte, returnGas uint64, err error) {
	if !evm.CanTransfer(callerAddr, chainId, value) {
		return nil, gas, ErrInsufficientBalance
	}

	byteCode := evm.State.GetByteCode(contractAddr, chainId)
	if byteCode == nil {
		return nil, gas, ErrCodeNotExists
	}
	evm.Transfer(callerAddr, contractAddr, chainId, value)

	contract := NewContract(callerAddr, gas, value, nil)
	contract.SetCode(contractAddr, byteCode)

	ret, err = run(evm, contract, input)
	return ret, contract.Gas, err
}

func (evm *EVM) StaticCall(callerAddr, contractAddr accounts.CommonAddress, chainId int64, input []byte, gas uint64) (ret []byte, returnGas uint64, err error) {

	byteCode := evm.State.GetByteCode(contractAddr, chainId)
	if byteCode == nil {
		return nil, gas, ErrCodeNotExists
	}

	contract := NewContract(callerAddr, gas, new(big.Int), nil)
	contract.SetCode(contractAddr, byteCode)

	ret, err = run(evm, contract, input)
	return ret, contract.Gas, err
}

func (evm *EVM) DelegateCall(con *Contract, contractAddr accounts.CommonAddress, input []byte, gas uint64) (ret []byte, leftGas uint64, err error) {

	callerAddr := con.CallerAddr
	chainId := con.ChainId
	jumpdests := con.Jumpdests

	byteCode := evm.State.GetByteCode(contractAddr, chainId)
	if byteCode == nil {
		return nil, gas, ErrCodeNotExists
	}

	contract := NewContract(callerAddr, gas, new(big.Int), jumpdests)
	contract.SetCode(contractAddr, byteCode)

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
	if interpreter.CanRun(contract.ByteCode) {
		return interpreter.Run(contract, input)
	}
	return nil, ErrNoCompatibleInterpreter
}

func (evm *EVM) CanTransfer(addr accounts.CommonAddress, chainId int64, amount *big.Int) bool {
	balance := evm.State.GetBalance(addr, chainId)
	return balance.Cmp(amount) >= 0
}

func (evm *EVM) Transfer(from, to accounts.CommonAddress, chainId int64, amount *big.Int) error {
	err := evm.State.SubBalance(from, chainId, amount)
	if err != nil {
		return err
	}
	err = evm.State.AddBalance(to, chainId, amount)
	if err != nil {
		evm.State.AddBalance(from, chainId, amount)
		return err
	}
	return nil
}


package vm

import (
	"math/big"
	"errors"
	"BlockChainTest/accounts"
	"fmt"
	"BlockChainTest/database"
	"BlockChainTest/config"
)

type EVM struct {
	State *State
	Interpreter *EVMInterpreter
	CallGasTemp uint64
	GasLimit uint64
	Origin accounts.CommonAddress
	GasPrice *big.Int
	CoinBase accounts.CommonAddress
	Time *big.Int
	Abort int32
}

var (
	ErrNoCompatibleInterpreter  = errors.New("no compatible interpreter")
)

func NewEVM(dt database.Transactional) *EVM {
	evm := &EVM{}
	evm.State = NewState(dt)
	evm.Interpreter = NewEVMInterpreter(evm)
	return evm
}

func (evm *EVM) CreateContractCode(callerAddr accounts.CommonAddress, chainId config.ChainIdType, byteCode accounts.ByteCode, gas uint64, value *big.Int) ([]byte, accounts.CommonAddress, uint64, error) {
	if !evm.CanTransfer(callerAddr, chainId, value) {
		return nil, accounts.CommonAddress{}, gas, ErrInsufficientBalance
	}

	nonce := evm.State.GetNonce(callerAddr, chainId) + 1
	account, err := evm.State.CreateContractAccount(callerAddr, chainId, nonce)
	if err != nil {
		return nil, accounts.CommonAddress{}, gas, err
	}

	contractAddr := account.Address
	evm.Transfer(callerAddr, contractAddr, chainId, value)

	fmt.Println("contract addr: ", contractAddr.Hex())

	contract := NewContract(callerAddr, chainId, gas, value, nil)
	contract.SetCode(contractAddr, byteCode)
	fmt.Println("contract gas: ", contract.Gas)

	ret, err := run(evm, contract, nil, false)
	if err != nil {
		return nil, accounts.CommonAddress{}, gas, err
	}

	err = evm.State.SetByteCode(contractAddr, chainId, ret)
	if err != nil {
		return nil, accounts.CommonAddress{}, gas, err
	}
	fmt.Println("contract address: ", contractAddr.Hex())
	fmt.Println("contract gas: ", contract.Gas)

	createDataGas := uint64(len(ret)) * CreateDataGas
	contract.UseGas(createDataGas)

	return ret, contractAddr, contract.Gas, nil
}

func (evm *EVM) CallContractCode(callerAddr, contractAddr accounts.CommonAddress, chainId config.ChainIdType, input []byte, gas uint64, value *big.Int) (ret []byte, returnGas uint64, err error) {
	if !evm.CanTransfer(callerAddr, chainId, value) {
		return nil, gas, ErrInsufficientBalance
	}

	byteCode := evm.State.GetByteCode(contractAddr, chainId)
	if byteCode == nil {
		return nil, gas, ErrCodeNotExists
	}
	evm.Transfer(callerAddr, contractAddr, chainId, value)

	contract := NewContract(callerAddr, chainId, gas, value, nil)
	contract.SetCode(contractAddr, byteCode)

	ret, err = run(evm, contract, input, false)
	return ret, contract.Gas, err
}

func (evm *EVM) StaticCall(callerAddr, contractAddr accounts.CommonAddress, chainId config.ChainIdType, input []byte, gas uint64) (ret []byte, returnGas uint64, err error) {

	byteCode := evm.State.GetByteCode(contractAddr, chainId)
	if byteCode == nil {
		return nil, gas, ErrCodeNotExists
	}

	contract := NewContract(callerAddr, chainId, gas, new(big.Int), nil)
	contract.SetCode(contractAddr, byteCode)

	ret, err = run(evm, contract, input, true)
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

	contract := NewContract(callerAddr, chainId, gas, new(big.Int), jumpdests)
	contract.SetCode(contractAddr, byteCode)

	ret, err = run(evm, contract, input, false)
	return ret, con.Gas, err
}


func run(evm *EVM, contract *Contract, input []byte, readOnly bool) ([]byte, error) {
	if !contract.ContractAddr.IsEmpty() {
		precompiles := PrecompiledContracts
		if p := precompiles[contract.ContractAddr]; p != nil {
			return RunPrecompiledContract(p, input, contract)
		}
	}
	interpreter := evm.Interpreter
	if interpreter.canRun(contract.ByteCode) {
		return interpreter.Run(contract, input, readOnly)
	}
	return nil, ErrNoCompatibleInterpreter
}

func (evm *EVM) CanTransfer(addr accounts.CommonAddress, chainId config.ChainIdType, amount *big.Int) bool {
	balance := evm.State.GetBalance(addr, chainId)
	return balance.Cmp(amount) >= 0
}

func (evm *EVM) Transfer(from, to accounts.CommonAddress, chainId config.ChainIdType, amount *big.Int) error {
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


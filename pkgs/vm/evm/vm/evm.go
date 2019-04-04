package vm

import (
	"errors"
	"fmt"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
)

type EVM struct {
	State       *State
	databaseApi *database.DatabaseService
	Interpreter *EVMInterpreter
	CallGasTemp uint64
	GasLimit    uint64
	Origin      crypto.CommonAddress
	GasPrice    *big.Int
	CoinBase    crypto.CommonAddress
	Time        *big.Int
	Abort       int32
	ChainId      app.ChainIdType
}

var (
	ErrNoCompatibleInterpreter  = errors.New("no compatible interpreter")
)

func NewEVM(databaseApi *database.DatabaseService, chainId app.ChainIdType) *EVM {
	evm := &EVM{}
	evm.State = NewState(databaseApi)
	evm.Interpreter = NewEVMInterpreter(evm)
	evm.ChainId = chainId
	return evm
}

func (evm *EVM) CreateContractCode(callerName string, contractName string, chainId app.ChainIdType, byteCode crypto.ByteCode, gas uint64, value *big.Int) ([]byte, string, uint64, error) {
	if !evm.CanTransfer(callerName, chainId, value) {
		return nil, "", gas, ErrInsufficientBalance
	}

	nonce := evm.State.GetNonce(callerName)
	_, err := evm.State.CreateContractAccount(contractName, evm.ChainId, nonce)
	if err != nil {
		return nil, "", gas, err
	}

	evm.Transfer(callerName, contractName, chainId, value)
	contract := NewContract(callerName, contractName, chainId, gas, value, nil)
	contract.SetCode(byteCode)
	ret, err := run(evm, contract, nil, false)

	createDataGas := uint64(len(ret)) * params.CreateDataGas
	if contract.UseGas(createDataGas) {
		err = evm.State.SetByteCode(contractName, ret)
	} else {
		err = ErrCodeStoreOutOfGas
	}

	if err != nil && err != ErrCodeStoreOutOfGas {
		//evm.State.dt.Discard()
		if err != errExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	} else {
		//evm.State.dt.Commit()
	}

	fmt.Println("contract address: ", contractName)
	fmt.Println("contract gas: ", contract.Gas)
	return ret, contractName, contract.Gas, err
}

func (evm *EVM) CallContractCode(callerName, contractName string, chainId app.ChainIdType, input []byte, gas uint64, value *big.Int) (ret []byte, returnGas uint64, err error) {
	if !evm.CanTransfer(callerName, chainId, value) {
		return nil, gas, ErrInsufficientBalance
	}

	byteCode := evm.State.GetByteCode(contractName)
	if byteCode == nil {
		return nil, gas, ErrCodeNotExists
	}

	evm.Transfer(callerName, contractName, chainId, value)
	contract := NewContract(callerName,contractName, chainId, gas, value, nil)
	contract.SetCode(byteCode)

	ret, err = run(evm, contract, input, false)
	if err != nil {
		//evm.State.dt.Discard()
		if err != errExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	} else {
		//evm.State.dt.Commit()
	}

	return ret, contract.Gas, err
}

func (evm *EVM) StaticCall(callerName, contractName string, chainId app.ChainIdType, input []byte, gas uint64) (ret []byte, returnGas uint64, err error) {
	byteCode := evm.State.GetByteCode(contractName)
	if byteCode == nil {
		return nil, gas, ErrCodeNotExists
	}

	contract := NewContract(callerName, contractName, chainId, gas, new(big.Int), nil)
	contract.SetCode(byteCode)

	ret, err = run(evm, contract, input, true)
	if err != nil {
		//evm.State.dt.Discard()
		if err != errExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	} else {
		//evm.State.dt.Commit()
	}

	return ret, contract.Gas, err
}

func (evm *EVM) DelegateCall(con *Contract, contractName string, input []byte, gas uint64) (ret []byte, leftGas uint64, err error) {
	callerName := con.CallerName
	chainId := con.ChainId
	jumpdests := con.Jumpdests

	byteCode := evm.State.GetByteCode(contractName)
	if byteCode == nil {
		return nil, gas, ErrCodeNotExists
	}

	contract := NewContract(callerName, contractName, chainId, gas, new(big.Int), jumpdests)
	contract.SetCode(byteCode)

	ret, err = run(evm, contract, input, false)
	if err != nil {
		//evm.State.dt.Discard()
		if err != errExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	} else {
		//evm.State.dt.Commit()
	}

	return ret, con.Gas, err
}


func run(evm *EVM, contract *Contract, input []byte, readOnly bool) ([]byte, error) {
	if contract.ContractName != "" {
		precompiles := PrecompiledContracts
		if p := precompiles[contract.ContractName]; p != nil {
			return RunPrecompiledContract(p, input, contract)
		}
	}
	interpreter := evm.Interpreter
	if interpreter.canRun(contract.ByteCode) {
		return interpreter.Run(contract, input, readOnly)
	}
	return nil, ErrNoCompatibleInterpreter
}

func (evm *EVM) CanTransfer(accountName string, chainId app.ChainIdType, amount *big.Int) bool {
	balance, err := evm.State.GetBalance(accountName)
	if err != nil {
		return false
	}
	return balance.Cmp(amount) >= 0
}

func (evm *EVM) Transfer(from, to string, chainId app.ChainIdType, amount *big.Int) error {
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


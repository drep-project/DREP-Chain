package core

import (
	"math/big"
	"BlockChainTest/bean"
	"fmt"
	"BlockChainTest/core/vm"
	"BlockChainTest/accounts"
	"bytes"
	"encoding/hex"
)

func ExecuteCreateCode(evm *vm.EVM, callerAddr accounts.CommonAddress, chainId int64, code []byte, gas uint64, value *big.Int) (uint64, error) {
	ret, _, returnGas, err := evm.CreateContractCode(callerAddr, chainId, code, gas, value)
	fmt.Println("gas: ", gas)
	fmt.Println("code: ", hex.EncodeToString(code))
	fmt.Println("ret: ", ret)
	fmt.Println("err: ", err)
	return returnGas, err
}

func ExecuteCallCode(evm *vm.EVM, callerAddr, contractAddr accounts.CommonAddress, chainId int64, input []byte, gas uint64, value *big.Int) (uint64, error) {
	ret, returnGas, err := evm.CallContractCode(callerAddr, contractAddr, chainId, input, gas, value)
	fmt.Println("ret: ", ret)
	fmt.Println("err: ", err)
	return returnGas, err
}

func ExecuteStaticCall(evm *vm.EVM, callerAddr, contractAddr accounts.CommonAddress, chainId int64, input []byte, gas uint64) (uint64, error) {
	ret, returnGas, err := evm.StaticCall(callerAddr, contractAddr, chainId, input, gas)
	fmt.Println("ret: ", ret)
	fmt.Println("err: ", err)
	return returnGas, err
}

type Message struct {
	From accounts.CommonAddress
	To accounts.CommonAddress
	ChainId int64
	DestChain int64
	Gas uint64
	Value *big.Int
	Nonce uint64
	Input []byte
	ReadOnly bool
}

func Tx2Message(tx *bean.Transaction) *Message {
	gasLimit := new(big.Int).SetBytes(tx.Data.GasLimit)
	gasValue := new(big.Int).SetBytes(tx.Data.GasPrice)
	gas := new(big.Int).Mul(gasLimit, gasValue)
	readOnly := false
	if bytes.Equal(tx.Data.Data[:1], []byte{1}) {
		readOnly = true
	}
	return &Message{
		From: accounts.PubKey2Address(tx.Data.PubKey),
		To: accounts.Hex2Address(tx.Data.To),
		ChainId: tx.Data.ChainId,
		DestChain: tx.Data.DestChain,
		Gas: gas.Uint64(),
		Value: new(big.Int).SetBytes(tx.Data.Amount),
		Nonce: uint64(tx.Data.Nonce),
		Input: tx.Data.Data[1:],
		ReadOnly: readOnly,
	}
}

func ApplyMessage(evm *vm.EVM, message *Message) (uint64, error) {
	contractCreation := message.To.IsEmpty()
	if contractCreation {
		return ExecuteCreateCode(evm, message.From, message.ChainId, message.Input, message.Gas, message.Value)
	} else if !message.ReadOnly {
		return ExecuteCallCode(evm, message.From, message.To, message.ChainId, message.Input, message.Gas, message.Value)
	} else {
		return ExecuteStaticCall(evm, message.From, message.To, message.ChainId, message.Input, message.Gas)
	}
}

func ApplyTransaction(evm *vm.EVM, tx *bean.Transaction) (uint64, error) {
	return ApplyMessage(evm, Tx2Message(tx))
}
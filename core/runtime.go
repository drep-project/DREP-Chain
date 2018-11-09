package _our

import (
	"math/big"
	"BlockChainTest/bean"
	"fmt"
	"BlockChainTest/core/vm"
)

func ExecuteCreateCode(callerAddr bean.CommonAddress, code []byte, gas uint64, value *big.Int) {
	evm := vm.NewEVM()
	ret, _, err := evm.CreateContractCode(callerAddr, code, gas, value)
	fmt.Println("ret: ", ret)
	fmt.Println("err: ", err)
}

func ExecuteCallCode(callerAddr, contractAddr bean.CommonAddress, input []byte, gas uint64, value *big.Int) {
	evm := vm.NewEVM()
	ret, _, err := evm.CallContractCode(callerAddr, contractAddr, input, gas, value)
	fmt.Println("ret: ", ret)
	fmt.Println("err: ", err)
}

type Message struct {
	From bean.CommonAddress
	To bean.CommonAddress
	Gas uint64
	Value *big.Int
	Nonce uint64
	Data []byte
}

func Tx2Message(tx *bean.Transaction) *Message {
	gasLimit := new(big.Int).SetBytes(tx.Data.GasLimit)
	gasValue := new(big.Int).SetBytes(tx.Data.GasPrice)
	gas := new(big.Int).Mul(gasLimit, gasValue)
	return &Message{
		From: tx.Address(),
		To: bean.Hex2Address(tx.Data.To),
		Gas: gas.Uint64(),
		Value: new(big.Int).SetBytes(tx.Data.Amount),
		Nonce: uint64(tx.Data.Nonce),
		Data: tx.Data.Data,
	}
}

func ApplyMessage(message *Message) {
	contractCreation := message.To.IsEmpty()
	if contractCreation {
		ExecuteCreateCode(message.From, message.Data, message.Gas, message.Value)
	} else {
		ExecuteCallCode(message.From, message.To, message.Data, message.Gas, message.Value)
	}
}

func ApplyTransaction(tx *bean.Transaction) {
	ApplyMessage(Tx2Message(tx))
}
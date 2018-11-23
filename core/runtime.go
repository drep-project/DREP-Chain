package _our

import (
	"math/big"
	"BlockChainTest/bean"
	"fmt"
	"BlockChainTest/core/vm"
	"BlockChainTest/accounts"
)

func ExecuteCreateCode(callerAddr accounts.CommonAddress, chainId int64, code []byte, gas uint64, value *big.Int) {
	evm := vm.NewEVM()
	ret, _, err := evm.CreateContractCode(callerAddr, chainId, code, gas, value)
	fmt.Println("ret: ", ret)
	fmt.Println("err: ", err)
}

func ExecuteCallCode(callerAddr, contractAddr accounts.CommonAddress, chainId int64, input []byte, gas uint64, value *big.Int) {
	evm := vm.NewEVM()
	ret, _, err := evm.CallContractCode(callerAddr, contractAddr, chainId, input, gas, value)
	fmt.Println("ret: ", ret)
	fmt.Println("err: ", err)
}

type Message struct {
	From accounts.CommonAddress
	To accounts.CommonAddress
	ChainId int64
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
		From: accounts.PubKey2Address(tx.Data.PubKey),
		To: accounts.Hex2Address(tx.Data.To),
		ChainId: tx.Data.ChainId,
		Gas: gas.Uint64(),
		Value: new(big.Int).SetBytes(tx.Data.Amount),
		Nonce: uint64(tx.Data.Nonce),
		Data: tx.Data.Data,
	}
}

func ApplyMessage(message *Message) {
	contractCreation := message.To.IsEmpty()
	if contractCreation {
		ExecuteCreateCode(message.From, message.ChainId, message.Data, message.Gas, message.Value)
	} else {
		ExecuteCallCode(message.From, message.To, message.ChainId, message.Data, message.Gas, message.Value)
	}
}

func ApplyTransaction(tx *bean.Transaction) {
	ApplyMessage(Tx2Message(tx))
}
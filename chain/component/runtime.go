package core

import (
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/chain/component/vm"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
)

func ExecuteCreateCode(evm *vm.EVM, callerAddr crypto.CommonAddress, chainId common.ChainIdType, code []byte, gas uint64, value *big.Int) (uint64, error) {
	ret, _, returnGas, err := evm.CreateContractCode(callerAddr, chainId, code, gas, value)
	fmt.Println("gas: ", gas)
	fmt.Println("code: ", hex.EncodeToString(code))
	fmt.Println("ret: ", ret)
	fmt.Println("err: ", err)
	return returnGas, err
}

func ExecuteCallCode(evm *vm.EVM, callerAddr, contractAddr crypto.CommonAddress, chainId common.ChainIdType, input []byte, gas uint64, value *big.Int) (uint64, error) {
	ret, returnGas, err := evm.CallContractCode(callerAddr, contractAddr, chainId, input, gas, value)
	fmt.Println("ret: ", ret)
	fmt.Println("err: ", err)
	return returnGas, err
}

func ExecuteStaticCall(evm *vm.EVM, callerAddr, contractAddr crypto.CommonAddress, chainId common.ChainIdType, input []byte, gas uint64) (uint64, error) {
	ret, returnGas, err := evm.StaticCall(callerAddr, contractAddr, chainId, input, gas)
	fmt.Println("ret: ", ret)
	fmt.Println("err: ", err)
	return returnGas, err
}

type Message struct {
	From      crypto.CommonAddress
	To        crypto.CommonAddress
	ChainId   common.ChainIdType
	DestChain common.ChainIdType
	Gas       uint64
	Value     *big.Int
	Nonce     uint64
	Input     []byte
	ReadOnly  bool
}

func Tx2Message(tx *chainTypes.Transaction) *Message {
	readOnly := false
	if bytes.Equal(tx.Data.Data[:1], []byte{1}) {
		readOnly = true
	}
	return &Message{
		From:      crypto.PubKey2Address(tx.Data.PubKey),
		To:        tx.Data.To,
		ChainId:   tx.Data.ChainId,
		Gas:       tx.Data.GasLimit.Uint64(),
		Value:     tx.Data.Amount,
		Nonce:     uint64(tx.Data.Nonce),
		Input:     tx.Data.Data[1:],
		ReadOnly:  readOnly,
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

func ApplyTransaction(evm *vm.EVM, tx *chainTypes.Transaction) (uint64, error) {
	return ApplyMessage(evm, Tx2Message(tx))
}
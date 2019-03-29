package lwasm

import (
	"errors"
	"fmt"
	"github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/database"
)

type WasmVm struct {
	databaseApi *database.DatabaseService
}

func  (wasmVm *WasmVm) executeCreateCode(message *types.Message, contractName string, code []byte) (uint64, error) {
	state := &State{
		databaseApi: wasmVm.databaseApi,
	}
	resolver := &Resolver{
		State   	:		state,
		Time     	:		uint64(message.Time),
		TxHash		:		message.TxHash,
		ChainId	:	message.ChainId,
		Code	:	code,
		ContractAccount :	contractName,
		CallerAccount	:	message.From,
	}
	runtime := &Runtime{
		ChainId			:	message.ChainId,
		Code			:	code,
		ContractAccount :	contractName,
		CallerAccount	:	message.From,
		State       	:	state,
		Resolve	  	    :	resolver,
	}
	gasUsed, err := runtime.Create()
	if err != nil {
		return message.Gas.Uint64(), err
	}
	fmt.Println(gasUsed)
	//TODO should to run init func after deploy contract
	return message.Gas.Uint64(), err
}

func  (wasmVm *WasmVm) executeCallCode(message *types.Message, contractName string, input []byte) (uint64, error) {
	storage, err := wasmVm.databaseApi.GetStorage(contractName, true)
	if err != nil {
		return 0, errors.New("contract not exist")
	}
	state := &State{
		databaseApi: wasmVm.databaseApi,
	}
	resolver := &Resolver{
		State   		:	state,
		Time     		:	uint64(message.Time),
		TxHash			:	message.TxHash,
		ChainId			:	message.ChainId,
		Code			:	storage.ByteCode,
		Input			:	input,
		ContractAccount :	contractName,
		CallerAccount	:	message.From,
	}
	runtime := &Runtime{
		ChainId			:	message.ChainId,
		Code			:	storage.ByteCode,
		Input 			:	input,
		ContractAccount :	contractName,
		CallerAccount	:	message.From,
		State       	:	state,
		Resolve	  		:	resolver,
		GasLimit		:	message.Gas.Uint64(),
	}
	_, err = runtime.Call()
	if err != nil {
		return message.Gas.Uint64(), err
	}

	return message.Gas.Uint64(), err
}

func  (wasmVm *WasmVm) RunMessage(message *types.Message) (uint64, error) {
	switch message.Type {
	case types.CreateContractType:
		createContractAction := message.Action.(*types.CreateContractAction)
		return  wasmVm.executeCreateCode(message, createContractAction.ContractName, createContractAction.ByteCode)
	case types.CallContractType:
		callContractAction := message.Action.(*types.CallContractAction)
		return  wasmVm.executeCallCode(message, callContractAction.ContractName,callContractAction.Input)
	}
	return 0, errors.New("not support tx type")
}
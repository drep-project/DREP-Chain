package lwasm

import (
	"errors"
	"github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/database"
)

type WasmVm struct {
	databaseApi *database.DatabaseService
}

func  (wasmVm *WasmVm) executeCreateCode(message *types.Message, contractName string, code []byte) ([]byte, uint64, error) {
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
		return nil, 0, err
	}
	//TODO should to run init func after deploy contract
	return nil, message.Gas.Uint64() - gasUsed, err
}

func  (wasmVm *WasmVm) executeCallCode(message *types.Message, contractName string, input []byte) ([]byte, uint64, error) {
	storage, err := wasmVm.databaseApi.GetStorage(contractName, true)
	if err != nil {
		return nil,0, errors.New("contract not exist")
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
	result, gasUsed, err := runtime.Call()
	if err != nil {
		return nil, 0, err
	}
	wasmVm.databaseApi.PutLogs(resolver.Logs, message.TxHash.Bytes())
	return result,  message.Gas.Uint64() - gasUsed, err
}

func  (wasmVm *WasmVm) RunMessage(message *types.Message) ([]byte, uint64, error) {
	switch message.Type {
	case types.CreateContractType:
		createContractAction := message.Action.(*types.CreateContractAction)
		return wasmVm.executeCreateCode(message, createContractAction.ContractName, createContractAction.ByteCode)
	case types.CallContractType:
		callContractAction := message.Action.(*types.CallContractAction)
		return  wasmVm.executeCallCode(message, callContractAction.ContractName,callContractAction.Input)
	}
	return nil, 0, errors.New("not support tx type")
}
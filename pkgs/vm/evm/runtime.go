package evm

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/pkgs/vm/evm/vm"
	"gopkg.in/urfave/cli.v1"
	"math/big"
)

var (
	DefaultEvmConfig = &VMConfig{}
)
func  (evmService *EvmService) ExecuteCreateCode(evm *vm.EVM, callerName, contractName string, chainId app.ChainIdType, code []byte, gas uint64, value *big.Int) (uint64, error) {
	ret, _, returnGas, err := evm.CreateContractCode(callerName, contractName, chainId, code, gas, value)
	fmt.Println("gas: ", gas)
	fmt.Println("code: ", hex.EncodeToString(code))
	fmt.Println("ret: ", ret)
	fmt.Println("err: ", err)
	return returnGas, err
}

func  (evmService *EvmService) ExecuteCallCode(evm *vm.EVM, callerName, contractName string, chainId app.ChainIdType, input []byte, gas uint64, value *big.Int) (uint64, error) {
	ret, returnGas, err := evm.CallContractCode(callerName, contractName, chainId, input, gas, value)
	fmt.Println("ret: ", ret)
	fmt.Println("err: ", err)
	return returnGas, err
}

func  (evmService *EvmService) ExecuteStaticCall(evm *vm.EVM, callerName, contractName string, chainId app.ChainIdType, input []byte, gas uint64) (uint64, error) {
	ret, returnGas, err := evm.StaticCall(callerName, contractName, chainId, input, gas)
	fmt.Println("ret: ", ret)
	fmt.Println("err: ", err)
	return returnGas, err
}


func  (evmService *EvmService) ApplyMessage(evm *vm.EVM, message *types.Message) (uint64, error) {
	switch message.Type {
	case types.CreateContractType:
		//createContractAction := message.Action.(*types.CreateContractAction)
	//	return  evmService.ExecuteCreateCode(evm, message.From, createContractAction.ContractName, message.ChainId, createContractAction.ByteCode, message.Gas.Uint64(), message.Value)
	case types.CallContractType:
	//	callContractAction := message.Action.(*types.CallContractAction)

	//	return  evmService.ExecuteStaticCall(evm, message.From, callContractAction.ContractName, message.ChainId, callContractAction.Input, message.Gas.Uint64())

		//return  evmService.ExecuteCallCode(evm, message.From, callContractAction.ContractName,message.ChainId, callContractAction.Input, message.Gas.Uint64(), message.Value)

	}
	return 0, errors.New("not support tx type")
}

type EvmService struct {
	Config *VMConfig
}

func (evmService *EvmService) Name() string {
	return "vm"
}

func (evmService *EvmService) Api() []app.API {
	return []app.API{}
}

func (evmService *EvmService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{}
}

func (evmService *EvmService) P2pMessages() map[int]interface{} {
	return map[int]interface{}{}
}

func (evmService *EvmService) Init(executeContext *app.ExecuteContext) error {
	evmService.Config = DefaultEvmConfig
	err := executeContext.UnmashalConfig(evmService.Name(), evmService.Config)
	if err != nil {
		return err
	}
	return nil
}

func (evmService *EvmService)  Start(executeContext *app.ExecuteContext) error {
	return nil
}

func (evmService *EvmService)  Stop(executeContext *app.ExecuteContext) error{
	return nil
}

func (evmService *EvmService)  Receive(context actor.Context) { }


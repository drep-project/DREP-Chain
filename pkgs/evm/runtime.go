package evm

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/pkgs/evm/vm"
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

func  (evmService *EvmService) ApplyMessage(evm *vm.EVM, tx *types.Transaction) (uint64, error) {
	switch tx.Type() {
	case types.CreateContractType:
		createContractAction :=	&types.CreateContractAction{}
		err := json.Unmarshal(tx.GetData(), createContractAction)
		if err != nil {
			return 0, err
		}
		return  evmService.ExecuteCreateCode(evm, tx.From(), createContractAction.ContractName, tx.ChainId(), createContractAction.ByteCode, tx.GasLimit().Uint64(), tx.Amount())
	case types.CallContractType:
		callContractAction := &types.CallContractAction{}
		err := json.Unmarshal(tx.GetData(), callContractAction)
		if err != nil {
			return 0, err
		}
		if callContractAction.Readonly {
			return  evmService.ExecuteStaticCall(evm, tx.From(), callContractAction.ContractName, tx.ChainId(), callContractAction.Input, tx.GasLimit().Uint64())
		}else{
			return  evmService.ExecuteCallCode(evm, tx.From(), callContractAction.ContractName, tx.ChainId(), callContractAction.Input, tx.GasLimit().Uint64(), tx.Amount())
		}
	}
	return 0, errors.New("not support tx type")
}

func  (evmService *EvmService) ApplyTransaction(evm *vm.EVM, tx *types.Transaction) (uint64, error) {
	return evmService.ApplyMessage(evm, tx)
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


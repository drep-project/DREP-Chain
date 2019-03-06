package evm

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/pkgs/evm/vm"
	"github.com/drep-project/drep-chain/transaction/types"
	"gopkg.in/urfave/cli.v1"
	"math/big"
)

var (
	DefaultEvmConfig = &VMConfig{}
)
func  (evmService *EvmService) ExecuteCreateCode(evm *vm.EVM, callerAddr crypto.CommonAddress, chainId app.ChainIdType, code []byte, gas uint64, value *big.Int) (uint64, error) {
	ret, _, returnGas, err := evm.CreateContractCode(callerAddr, chainId, code, gas, value)
	fmt.Println("gas: ", gas)
	fmt.Println("code: ", hex.EncodeToString(code))
	fmt.Println("ret: ", ret)
	fmt.Println("err: ", err)
	return returnGas, err
}

func  (evmService *EvmService) ExecuteCallCode(evm *vm.EVM, callerAddr, contractAddr crypto.CommonAddress, chainId app.ChainIdType, input []byte, gas uint64, value *big.Int) (uint64, error) {
	ret, returnGas, err := evm.CallContractCode(callerAddr, contractAddr, chainId, input, gas, value)
	fmt.Println("ret: ", ret)
	fmt.Println("err: ", err)
	return returnGas, err
}

func  (evmService *EvmService) ExecuteStaticCall(evm *vm.EVM, callerAddr, contractAddr crypto.CommonAddress, chainId app.ChainIdType, input []byte, gas uint64) (uint64, error) {
	ret, returnGas, err := evm.StaticCall(callerAddr, contractAddr, chainId, input, gas)
	fmt.Println("ret: ", ret)
	fmt.Println("err: ", err)
	return returnGas, err
}

func  (evmService *EvmService) Tx2Message(tx *types.Transaction) *Message {
	readOnly := false
	if bytes.Equal(tx.Data()[:1], []byte{1}) {
		readOnly = true
	}

	return &Message{
		From:      *tx.From(),
		To:        *tx.To(),
		ChainId:   tx.ChainId(),
		Gas:       tx.GasLimit().Uint64(),
		Value:     tx.Amount(),
		Nonce:     uint64(tx.Nonce()),
		Input:     tx.Data()[1:],
		ReadOnly:  readOnly,
	}
}

func  (evmService *EvmService) ApplyMessage(evm *vm.EVM, message *Message) (uint64, error) {
	contractCreation := message.To.IsEmpty()
	if contractCreation {
		return  evmService.ExecuteCreateCode(evm, message.From, message.ChainId, message.Input, message.Gas, message.Value)
	} else if !message.ReadOnly {
		return  evmService.ExecuteCallCode(evm, message.From, message.To, message.ChainId, message.Input, message.Gas, message.Value)
	} else {
		return  evmService.ExecuteStaticCall(evm, message.From, message.To, message.ChainId, message.Input, message.Gas)
	}
}

func  (evmService *EvmService) ApplyTransaction(evm *vm.EVM, tx *types.Transaction) (uint64, error) {
	return evmService.ApplyMessage(evm,  evmService.Tx2Message(tx))
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


package evm

import (
	"math/big"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/DREP-Chain/app"
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/database"
	"github.com/drep-project/DREP-Chain/params"
	"github.com/drep-project/DREP-Chain/pkgs/evm/vm"
	"github.com/drep-project/DREP-Chain/types"
	"github.com/drep-project/dlog"
	"gopkg.in/urfave/cli.v1"
)

var (
	DefaultEvmConfig = &vm.VMConfig{
		LogConfig: &vm.LogConfig{
			DisableMemory:  false,
			DisableStack:   false,
			DisableStorage: false,
			Debug:          false,
			Limit:          0,
		},
	}
)

type EvmService struct {
	Config *vm.VMConfig
	//Chain           chain.ChainServiceInterface `service:"chain"`
	DatabaseService *database.DatabaseService `service:"database"`
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
	//evmService.Chain.AddTransactionValidator(&EvmDeployTransactionSelector{}, &EvmDeployTransactionExecutor{evmService})
	return nil
}

func (evmService *EvmService) Start(executeContext *app.ExecuteContext) error {
	return nil
}

func (evmService *EvmService) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

func (evmService *EvmService) Receive(context actor.Context) {}

func (evmService *EvmService) Call(database store.StoreInterface, tx *types.Transaction, header *types.BlockHeader) (ret []byte, err error) {
	state := vm.NewState(database, header.Height)
	sender, err := tx.From()
	if err != nil {
		return nil, err
	}

	// Create a new context to be used in the EVM environment
	context := NewEVMContext(tx, header, sender)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewEVM(context, state, evmService.Config)

	ret, _, vmerr := vmenv.Call(*sender, *tx.To(), vmenv.ChainId, tx.Data.Data, tx.Gas(), tx.Amount())
	if vmerr != nil {
		dlog.Debug("call VM returned with error", "err", vmerr)
		return nil, vmerr
	}
	return ret, nil
}

func (evmService *EvmService) Eval(state vm.VMState, tx *types.Transaction, header *types.BlockHeader, gas uint64, value *big.Int) (ret []byte, gasUsed uint64, contractAddr crypto.CommonAddress, failed bool, err error) {
	sender, err := tx.From()
	if err != nil {
		return nil, uint64(0), crypto.CommonAddress{}, false, err
	}
	contractCreation := (tx.To() == nil || tx.To().IsEmpty()) && tx.Type() == types.CreateContractType

	// Create a new context to be used in the EVM environment
	context := NewEVMContext(tx, header, sender)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewEVM(context, state, evmService.Config)
	var (
		// vm errors do not effect consensus and are therefor
		// not assigned to err, except for insufficient balance
		// error.
		vmerr error
	)
	if contractCreation {
		ret, contractAddr, gas, vmerr = vmenv.Create(*sender, tx.Data.Data, gas, value)
	} else {
		// Increment the nonce for the next transaction
		state.SetNonce(sender, state.GetNonce(sender)+1)
		ret, gas, vmerr = vmenv.Call(*sender, *tx.To(), vmenv.ChainId, tx.Data.Data, gas, value)
	}
	if vmerr != nil {
		dlog.Debug("VM returned with error", "err", vmerr)
		// The only possible consensus-error would be if there wasn't
		// sufficient balance to make the transfer happen. The first
		// balance transfer may never fail.
		if vmerr == vm.ErrInsufficientBalance {
			return nil, uint64(0), crypto.CommonAddress{}, false, vmerr
		}
	}

	return ret, gas, contractAddr, vmerr != nil, err
}

func (evmService *EvmService) DefaultConfig(netType params.NetType) *vm.VMConfig {
	return DefaultEvmConfig
}

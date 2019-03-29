package lwasm

import (
	"bytes"
	"errors"
	"github.com/drep-project/drep-chain/app"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/go-interpreter/wagon/validate"
	"github.com/go-interpreter/wagon/wasm"
	"github.com/perlin-network/life/exec"
)

type Runtime struct {
	ChainId	app.ChainIdType
	Input			[]byte
	Abort       	int32
	Code			[]byte
	ContractAccount string
	CallerAccount	string
	State       	*State
	Resolve	    	*Resolver
	GasLimit    	uint64
}

func (runtime *Runtime) Create()  (uint64, error) {
	//TODO how to cacl deploy fee may bee a fix value
	newModule, err := wasm.ReadModule(bytes.NewReader(runtime.Code), nil)
	if err != nil {
		return runtime.GasLimit, err
	}

	err = validate.VerifyModule(newModule)
	if err != nil {
		return runtime.GasLimit, err
	}

	account, err := chainTypes.NewContractAccount(runtime.ContractAccount, runtime.ChainId)
	if err != nil {
		return runtime.GasLimit, err
	}
	account.Storage.ByteCode  = runtime.Code
	runtime.State.databaseApi.PutStorage(account.Name, account.Storage, true)
	return 0, err
}

func (runtime *Runtime) Call() (uint64, error) {
	vm, err := exec.NewVirtualMachine(runtime.Code, exec.VMConfig{
		DefaultMemoryPages:   128,
		DefaultTableSize:     65536,
		GasLimit: 			  runtime.GasLimit,
	}, runtime.Resolve, &SimpleGasPolicy{1})

	if err != nil {
		return runtime.GasLimit, err
	}

	// Get the function ID of the entry function to be executed.
	entryID, ok := vm.GetFunctionExport("invoke")
	if !ok {
		return runtime.GasLimit, errors.New("invoke not exist")
	}
	_, err = vm.RunWithGasLimit(entryID, runtime.GasLimit)
	if err != nil {
		vm.PrintStackTrace()
		return runtime.GasLimit, err
	}
	return vm.Gas, err
}
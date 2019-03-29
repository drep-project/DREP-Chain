package dwasm

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/drep-project/drep-chain/app"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/validate"
	"github.com/go-interpreter/wagon/wasm"
	"reflect"
)

type Runtime struct {
	ChainId	app.ChainIdType
	Input			[]byte
	Output			[]byte
	Abort       	int32
	Code			[]byte
	ContractAccount string
	CallerAccount	string

	State       	*State
	ImportFunc	    *ImportFunc
}

func (runtime *Runtime) Create()  error {
	account, err := chainTypes.NewContractAccount(runtime.ContractAccount, runtime.ChainId)
	if err != nil {
		return err
	}
	account.Storage.ByteCode  = runtime.Code
	runtime.State.databaseApi.PutStorage(account.Name, account.Storage, true)
	//TODO should to run init func after deploy contract
	return nil
}

func (runtime *Runtime) Call() ([]byte, error) {
	codeModule, err := wasm.ReadModule(bytes.NewReader(runtime.Code), runtime.Importer)
	if err != nil {
		return nil, err
	}

	err = validate.VerifyModule(codeModule)
	if err != nil {
		return nil, err
	}

	if codeModule.Export == nil {
		return nil, err
	}

	vm, err := exec.NewVM(codeModule)
	vm.RecoverPanic = true
	if err != nil {
		return nil, err
	}
	exportEntry := codeModule.Export.Entries["invoke"]
	_, err = vm.ExecCode(int64(exportEntry.Index))
	if err != nil {
		return nil, err
	}

	return runtime.Output, err
}

func (runtime *Runtime) Importer(name string) (*wasm.Module, error) {
	switch name {
	case "env":
		return runtime.EnvImport(), nil
	default:
		return nil, errors.New("not support module")
	}
}

func (runtime *Runtime) EnvImport() *wasm.Module {
	newModule := wasm.NewModule()
	importTypes := &wasm.SectionTypes{
		Entries: []wasm.FunctionSig{
			//func()uint64    [0]
			{
				Form:        0, // value for the 'func' type constructor
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI64},
			},
			//func()uint32     [1]
			{
				Form:        0, // value for the 'func' type constructor
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32)     [2]
			{
				Form:       0, // value for the 'func' type constructor
				ParamTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32)uint32  [3]
			{
				Form:        0, // value for the 'func' type constructor
				ParamTypes:  []wasm.ValueType{wasm.ValueTypeI32},
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32,uint32)  [4]
			{
				Form:       0, // value for the 'func' type constructor
				ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32},
			},
			//func(uint32,uint32,uint32)uint32  [5]
			{
				Form:        0, // value for the 'func' type constructor
				ParamTypes:  []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32},
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32,uint32,uint32,uint32,uint32)uint32  [6]
			{
				Form:        0, // value for the 'func' type constructor
				ParamTypes:  []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32},
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32,uint32,uint32,uint32)  [7]
			{
				Form:       0, // value for the 'func' type constructor
				ParamTypes: []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32, wasm.ValueTypeI32},
			},
			//func(uint32,uint32)uint32   [8]
			{
				Form:        0, // value for the 'func' type constructor
				ParamTypes:  []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32},
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//func(uint32,uint32)uint32   [9]
			{
				Form:        0, // value for the 'func' type constructor
				ParamTypes:  []wasm.ValueType{wasm.ValueTypeI32, wasm.ValueTypeI32,wasm.ValueTypeI32, wasm.ValueTypeI32},
				ReturnTypes: []wasm.ValueType{wasm.ValueTypeI32},
			},
			//funct()   [10]
			{
				Form: 0, // value for the 'func' type constructor
			},
		},
	}
	newModule.FunctionIndexSpace = []wasm.Function{
		{ //0
			Sig:  &importTypes.Entries[0],
			Host: reflect.ValueOf(runtime.ImportFunc.TimeStamp),
			Body: &wasm.FunctionBody{},
		},
		{ //1
			Sig:  &importTypes.Entries[3],
			Host: reflect.ValueOf(runtime.ImportFunc.GetTxHash),
			Body: &wasm.FunctionBody{},
		},
		{ //2
			Sig:  &importTypes.Entries[1],
			Host: reflect.ValueOf(runtime.GetContractAccountLength),
			Body: &wasm.FunctionBody{},
		},
		{ //3
			Sig:  &importTypes.Entries[2],
			Host: reflect.ValueOf(runtime.GetContractAccount),
			Body: &wasm.FunctionBody{},
		},
		{ //4
			Sig:  &importTypes.Entries[1],
			Host: reflect.ValueOf(runtime.GetCallerAccountLength),
			Body: &wasm.FunctionBody{},
		},
		{ //5
			Sig:  &importTypes.Entries[2],
			Host: reflect.ValueOf(runtime.GetCallerAccount),
			Body: &wasm.FunctionBody{},
		},
		{ //6
			Sig:  &importTypes.Entries[1],
			Host: reflect.ValueOf(runtime.GetInputLength),
			Body: &wasm.FunctionBody{},
		},
		{ //7
			Sig:  &importTypes.Entries[2],
			Host: reflect.ValueOf(runtime.GetInput),
			Body: &wasm.FunctionBody{},
		},
		{ //8
			Sig:  &importTypes.Entries[1],
			Host: reflect.ValueOf(runtime.GetCallOutputLength),
			Body: &wasm.FunctionBody{},
		},
		{ //9
			Sig:  &importTypes.Entries[2],
			Host: reflect.ValueOf(runtime.GetCallOutput),
			Body: &wasm.FunctionBody{},
		},
		{ //10
			Sig:  &importTypes.Entries[4],
			Host: reflect.ValueOf(runtime.Ret),
			Body: &wasm.FunctionBody{},
		},
		{ //11
			Sig:  &importTypes.Entries[4],
			Host: reflect.ValueOf(runtime.ImportFunc.Notify),
			Body: &wasm.FunctionBody{},
		},
		{ //12
			Sig:  &importTypes.Entries[5],
			Host: reflect.ValueOf(runtime.CallContract),
			Body: &wasm.FunctionBody{},
		},
		{ //13
			Sig:  &importTypes.Entries[6],
			Host: reflect.ValueOf(runtime.ImportFunc.StorageRead),
			Body: &wasm.FunctionBody{},
		},
		{ //14
			Sig:  &importTypes.Entries[7],
			Host: reflect.ValueOf(runtime.ImportFunc.StorageWrite),
			Body: &wasm.FunctionBody{},
		},
		{ //15
			Sig:  &importTypes.Entries[4],
			Host: reflect.ValueOf(runtime.ImportFunc.StorageDelete),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //16
			Sig:  &importTypes.Entries[6],
			Host: reflect.ValueOf(runtime.ImportFunc.GetBalance),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //17
			Sig:  &importTypes.Entries[6],
			Host: reflect.ValueOf(runtime.ImportFunc.GetReputation),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //18
			Sig:  &importTypes.Entries[8],
			Host: reflect.ValueOf(runtime.ImportFunc.ValidateAccount),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
		{ //19
			Sig:  &importTypes.Entries[4],
			Host: reflect.ValueOf(runtime.ImportFunc.Debug),
			Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
		},
	}

	newModule.Export = &wasm.SectionExports{
		Entries: map[string]wasm.ExportEntry{
			"timestamp": {
				FieldStr: "timestamp",
				Kind:     wasm.ExternalFunction,
				Index:    0,
			},
			"current_txhash": {
				FieldStr: "current_txhash",
				Kind:     wasm.ExternalFunction,
				Index:    1,
			},
			"contract_account_len": {
				FieldStr: "contract_account_len",
				Kind:     wasm.ExternalFunction,
				Index:    2,
			},
			"contract_account": {
				FieldStr: "contract_account",
				Kind:     wasm.ExternalFunction,
				Index:    3,
			},
			"caller_account_len": {
				FieldStr: "caller_account_len",
				Kind:     wasm.ExternalFunction,
				Index:    4,
			},
			"caller_account": {
				FieldStr: "caller_account",
				Kind:     wasm.ExternalFunction,
				Index:    5,
			},
			"input_length": {
				FieldStr: "input_length",
				Kind:     wasm.ExternalFunction,
				Index:    6,
			},
			"get_input": {
				FieldStr: "get_input",
				Kind:     wasm.ExternalFunction,
				Index:    7,
			},
			"call_output_length": {
				FieldStr: "call_output_length",
				Kind:     wasm.ExternalFunction,
				Index:    8,
			},
			"get_call_output": {
				FieldStr: "get_call_output",
				Kind:     wasm.ExternalFunction,
				Index:    9,
			},
			"ret": {
				FieldStr: "ret",
				Kind:     wasm.ExternalFunction,
				Index:    10,
			},
			"notify": {
				FieldStr: "notify",
				Kind:     wasm.ExternalFunction,
				Index:    11,
			},
			"call_contract": {
				FieldStr: "call_contract",
				Kind:     wasm.ExternalFunction,
				Index:    12,
			},
			"storage_read": {
				FieldStr: "storage_read",
				Kind:     wasm.ExternalFunction,
				Index:    13,
			},
			"storage_write": {
				FieldStr: "storage_write",
				Kind:     wasm.ExternalFunction,
				Index:    14,
			},
			"storage_delete": {
				FieldStr: "storage_delete",
				Kind:     wasm.ExternalFunction,
				Index:    15,
			},
			"get_balance": {
				FieldStr: "get_balance",
				Kind:     wasm.ExternalFunction,
				Index:    16,
			},
			"get_reputation": {
				FieldStr: "get_reputation",
				Kind:     wasm.ExternalFunction,
				Index:    17,
			},
			"validate_account": {
				FieldStr: "validate_account",
				Kind:     wasm.ExternalFunction,
				Index:    18,
			},
			"debug": {
				FieldStr: "debug",
				Kind:     wasm.ExternalFunction,
				Index:    19,
			},
		},
	}

	return newModule
}

func (runtime *Runtime) GetInputLength(proc *exec.Process) uint32{
	return uint32(len(runtime.Input))
}

func (runtime *Runtime) GetInput(proc *exec.Process, ptr uint32){
	_, err := proc.WriteAt(runtime.Input, int64(ptr))
	if err != nil {
		panic(err)
	}
}

func (runtime *Runtime) GetCallOutputLength(proc *exec.Process) uint32 {
	return uint32(len(runtime.Output))
}

func (runtime *Runtime) GetCallOutput(proc *exec.Process, ptr uint32) {
	_, err := proc.WriteAt(runtime.Output, int64(ptr))
	if err != nil {
		panic(err)
	}
}

func (runtime *Runtime) Ret(proc *exec.Process, ptr uint32, len uint32) {
	bs := make([]byte, len)
	_, err := proc.ReadAt(bs, int64(ptr))
	if err != nil {
		panic(err)
	}

	runtime.Output = bs
	proc.Terminate()
	fmt.Println(string(bs))
}


// ------------------------------Contract Message--------------------------

func (runtime *Runtime) GetContractAccountLength(proc *exec.Process) uint32{
	return uint32(len(runtime.ContractAccount))
}

func (runtime *Runtime) GetContractAccount(proc *exec.Process, ptr uint32){
	_, err := proc.WriteAt([]byte(runtime.ContractAccount), int64(ptr))
	if err != nil {
		panic(err)
	}
}

func (runtime *Runtime) GetCallerAccountLength(proc *exec.Process) uint32{
	return uint32(len(runtime.CallerAccount))
}

func (runtime *Runtime) GetCallerAccount(proc *exec.Process, ptr uint32){
	_, err := proc.WriteAt([]byte(runtime.CallerAccount), int64(ptr))
	if err != nil {
		panic(err)
	}
}

func (runtime *Runtime) CallContract(proc *exec.Process, dst uint32) {

}
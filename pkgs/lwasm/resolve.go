package lwasm

import (
	"errors"
	"fmt"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/perlin-network/life/exec"
	"github.com/drep-project/drep-chain/crypto"
	"io"
	"math"
)

type Resolver struct {
	//tempRet0 int64
	State       	*State
	Time        	uint64
	TxHash			crypto.Hash
	ChainId	app.ChainIdType
	Input			[]byte
	Output			[]byte
	CallOutput		[]byte
	Abort       	int32
	Code			[]byte
	ContractAccount string
	CallerAccount	string
	Logs            []types.Log
}

const (
	TIME_GAS 					= 10
	TXHASH_GAS 					= 10
	CONTRACT_ACCOOUNT_LEN_GAS	= 10
	CONTRACT_ACCOOUNT_GAS 		= 10
	CALLER_ACCOOUNT_LEN_GAS 	= 10
	CALLER_ACCOOUNT_GAS 		= 10
	INPUT_LEN_GAS 				= 10
	INPUT_GAS 					= 10
	CALL_OUTPUT_LEN_GAS 		= 10
	CALL_OUTPUT_GAS				= 10
	RET_GAS						= 10
	NOTIFY_GAS					= 10
	CALL_CONTRACT_GAS			= 10000
	GET_STORAGE_GAS				= 10
	WRITE_STORAGE_GAS			= 10
	DEL_STORAGE_GAS				= 10
	GET_BALANCE_GAS				= 10
	GET_REPUTATION_GAS			= 10
	VALIDATE_ACCOUNT_GAS		= 10
	DEBUG_GAS					= 10
)
func (r *Resolver) Copy() *Resolver {
	return &Resolver{
		State     	  	:	r.State,
		Time         	:	r.Time,
		TxHash		  	:	r.TxHash,
		ChainId   		:	r.ChainId,
		Abort       	  :	r.Abort,
		Code			  :	r.Code,
		ContractAccount   :	r.ContractAccount,
		CallerAccount	  :	r.CallerAccount,
	}
}
func (r *Resolver) ResolveGlobal(module, field string) int64 {
	panic("global import not allowed")
}
// ResolveFunc defines a set of import functions that may be called within a WebAssembly module.
func (r *Resolver) ResolveFunc(module, field string) exec.FunctionImport {
	switch module {
	case "env":
		switch field {
		case "timestamp":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(TIME_GAS) {
					vm.GasLimitExceeded = true
				}
				return int64(r.Time)
			}
		case "current_txhash":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(TXHASH_GAS) {
					vm.GasLimitExceeded = true
				}
				ptr := vm.GetCurrentFrame().Locals[0]

				length, err := r.WriteAt(vm.Memory, r.TxHash[:], int64(ptr))
				if err != nil {
					panic(err)
				}

				return int64(length)
			}
		case "contract_account_len":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(CONTRACT_ACCOOUNT_LEN_GAS) {
					vm.GasLimitExceeded = true
				}
				return int64(len(r.ContractAccount))
			}
		case "contract_account":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(CONTRACT_ACCOOUNT_GAS) {
					vm.GasLimitExceeded = true
				}
				ptr := vm.GetCurrentFrame().Locals[0]
				length, err := r.WriteAt(vm.Memory, []byte(r.ContractAccount), int64(ptr))
				if err != nil {
					panic(err)
				}

				return int64(length)
			}
		case "caller_account_len":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(CALLER_ACCOOUNT_LEN_GAS) {
					vm.GasLimitExceeded = true
				}
				return int64(len(r.CallerAccount))
			}
		case "caller_account":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(CALLER_ACCOOUNT_GAS) {
					vm.GasLimitExceeded = true
				}
				ptr := vm.GetCurrentFrame().Locals[0]
				length, err := r.WriteAt(vm.Memory, []byte(r.CallerAccount), int64(ptr))
				if err != nil {
					panic(err)
				}

				return int64(length)
			}
		case "input_length":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(INPUT_LEN_GAS) {
					vm.GasLimitExceeded = true
				}
				return int64(len(r.Input))
			}
		case "get_input":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(INPUT_GAS) {
					vm.GasLimitExceeded = true
				}
				ptr := vm.GetCurrentFrame().Locals[0]
				length, err := r.WriteAt(vm.Memory, r.Input, int64(ptr))
				if err != nil {
					panic(err)
				}
				return int64(length)
			}
		case "call_output_length":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(CALL_OUTPUT_LEN_GAS) {
					vm.GasLimitExceeded = true
				}
				return int64(len(r.CallOutput))
			}
		case "get_call_output":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(CALL_OUTPUT_GAS) {
					vm.GasLimitExceeded = true
				}
				ptr := vm.GetCurrentFrame().Locals[0]
				length, err := r.WriteAt(vm.Memory, r.CallOutput, int64(ptr))
				if err != nil {
					panic(err)
				}
				return int64(length)
			}
		case "ret":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(RET_GAS) {
					vm.GasLimitExceeded = true
				}
				ptr := vm.GetCurrentFrame().Locals[0]
				len := vm.GetCurrentFrame().Locals[1]
				bs := make([]byte, len)
				_, err := r.ReadAt(vm.Memory, bs, int64(ptr))
				if err != nil {
					panic(err)
				}

				r.Output = bs
				vm.Exited = true
				fmt.Println(string(bs))
				return 0
			}
		case "notify":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(NOTIFY_GAS) {
					vm.GasLimitExceeded = true
				}
				ptr := vm.GetCurrentFrame().Locals[0]
				len := vm.GetCurrentFrame().Locals[1]
				msg := make([]byte, len)
				_, err := r.ReadAt(vm.Memory, msg, int64(ptr))
				if err != nil {
					panic(err)
				}
				fmt.Println(string(msg))
				log :=types.Log{
					Name:    r.ContractAccount,
					ChainId: r.ChainId,
					TxHash:  r.TxHash.Bytes(),
					//Topics: r.ContractAccount,
					Data: 		msg,
				}
				r.Logs = append(r.Logs, log)
				return 0
			}
		case "call_contract":
			return func(vm *exec.VirtualMachine) int64 {
				//TODO CALL
				accountPtr := vm.GetCurrentFrame().Locals[0]
				accountLen := vm.GetCurrentFrame().Locals[1]
				codePtr := vm.GetCurrentFrame().Locals[0]
				codeLen := vm.GetCurrentFrame().Locals[1]
				if !vm.AddAndCheckGas(CALL_CONTRACT_GAS) {
					vm.GasLimitExceeded = true
				}

				account := make([]byte, accountLen)
				_, err := r.ReadAt(vm.Memory, account, int64(accountPtr))
				if err != nil {
					panic(err)
				}

				code := make([]byte, codeLen)
				_, err = r.ReadAt(vm.Memory, code, int64(codePtr))
				if err != nil {
					panic(err)
				}
				storage, err := r.State.databaseApi.GetStorage(string(account), true)
				if err != nil {
					panic(err)
				}
				newResolve := r.Copy()
				newResolve.Code =  storage.ByteCode
				newResolve.Input = code
				newVm, err := exec.NewVirtualMachine(code, exec.VMConfig{
					DefaultMemoryPages:   128,
					DefaultTableSize:     65536,
					GasLimit: 			  vm.Config.GasLimit,
				}, newResolve, &SimpleGasPolicy{1})
				newVm.Gas = vm.Gas

				if err != nil {
					panic(err)
				}
			
				// Get the function ID of the entry function to be executed.
				entryID, ok := newVm.GetFunctionExport("invoke")
				if !ok {
					panic(errors.New("invoke not exist"))
				}
				_, err = newVm.RunWithGasLimit(entryID, vm.Config.GasLimit)
				if err != nil {
					vm.PrintStackTrace()
					panic(err)
				}
				r.CallOutput = newResolve.Output
				return 0
			}
		case "storage_read":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(GET_STORAGE_GAS) {
					vm.GasLimitExceeded = true
				}
				keyPtr := vm.GetCurrentFrame().Locals[0]
				klen := vm.GetCurrentFrame().Locals[1]
				vPtr := vm.GetCurrentFrame().Locals[2]
				vlen := vm.GetCurrentFrame().Locals[3]
				offset := vm.GetCurrentFrame().Locals[4]
				keybytes := make([]byte, klen)
				_, err := r.ReadAt(vm.Memory,keybytes, int64(keyPtr))
				if err != nil {
					panic(err)
				}


				item := r.State.Load(sha3.HashS256([]byte(r.ContractAccount), keybytes))
				if item == nil {
					return math.MaxUint32
				}
				length := vlen
				itemlen := int64(len(item))
				if itemlen < vlen {
					length = itemlen
				}

				if int64(len(item)) < offset {
					panic(errors.New("offset is invalid"))
				}
				_, err = r.WriteAt(vm.Memory,item[offset:offset+length], int64(vPtr))

				if err != nil {
					panic(err)
				}
				return int64(len(item))
			}
		case "storage_write":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(WRITE_STORAGE_GAS) {
					vm.GasLimitExceeded = true
				}
				keyPtr := vm.GetCurrentFrame().Locals[0]
				keylen := vm.GetCurrentFrame().Locals[1]
				valPtr := vm.GetCurrentFrame().Locals[2]
				valLen := vm.GetCurrentFrame().Locals[3]
				keybytes := make([]byte, keylen)
				_, err := r.ReadAt(vm.Memory, keybytes, int64(keyPtr))
				if err != nil {
					panic(err)
				}

				valbytes := make([]byte, valLen)
				_, err = r.ReadAt(vm.Memory, valbytes, int64(valPtr))
				if err != nil {
					panic(err)
				}

				modifiedLoc := sha3.HashS256([]byte(r.ContractAccount), keybytes)
				r.State.Store(modifiedLoc, sha3.HashS256(valbytes))
				return 0
			}
		case "storage_delete":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(DEL_STORAGE_GAS) {
					vm.GasLimitExceeded = true
				}
				keyPtr := vm.GetCurrentFrame().Locals[0]
				keylen := vm.GetCurrentFrame().Locals[1]
				keybytes := make([]byte, keylen)
				_, err := r.ReadAt(vm.Memory, keybytes, int64(keyPtr))
				if err != nil {
					panic(err)
				}
				r.State.Delete(keybytes)
				return 0
			}
		case "get_balance":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(GET_BALANCE_GAS) {
					vm.GasLimitExceeded = true
				}
				keyPtr := vm.GetCurrentFrame().Locals[0]
				klen := vm.GetCurrentFrame().Locals[1]
				vPtr := vm.GetCurrentFrame().Locals[2]
				vlen := vm.GetCurrentFrame().Locals[3]
				offset := vm.GetCurrentFrame().Locals[4]
				keybytes := make([]byte, klen)
				_, err := r.ReadAt(vm.Memory, keybytes, int64(keyPtr))
				if err != nil {
					panic(err)
				}

				balance := r.State.GetBalance(string(keybytes)).Bytes()
				if balance == nil {
					return math.MaxUint32
				}
				length := vlen
				itemlen := int64(len(balance))
				if itemlen < vlen {
					length = itemlen
				}

				if int64(len(balance)) < offset {
					panic(errors.New("offset is invalid"))
				}
				_, err = r.WriteAt(vm.Memory, balance[offset:offset+length], int64(vPtr))

				if err != nil {
					panic(err)
				}
				return int64(len(balance))
			}
		case "get_reputation":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(GET_REPUTATION_GAS) {
					vm.GasLimitExceeded = true
				}
				keyPtr := vm.GetCurrentFrame().Locals[0]
				klen := vm.GetCurrentFrame().Locals[1]
				vPtr := vm.GetCurrentFrame().Locals[2]
				vlen := vm.GetCurrentFrame().Locals[3]
				offset := vm.GetCurrentFrame().Locals[4]
				keybytes := make([]byte, klen)
				_, err := r.ReadAt(vm.Memory, keybytes, int64(keyPtr))
				if err != nil {
					panic(err)
				}

				reputation := r.State.GetReputation(string(keybytes)).Bytes()
				if reputation == nil {
					return math.MaxUint32
				}
				length := vlen
				itemlen := int64(len(reputation))
				if itemlen < vlen {
					length = itemlen
				}

				if int64(len(reputation)) < offset {
					panic(errors.New("offset is invalid"))
				}
				_, err = r.WriteAt(vm.Memory, reputation[offset:offset+length], int64(vPtr))

				if err != nil {
					panic(err)
				}
				return int64(len(reputation))
			}
		case "validate_account":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(VALIDATE_ACCOUNT_GAS) {
					vm.GasLimitExceeded = true
				}
				ptr := vm.GetCurrentFrame().Locals[0]
				len := vm.GetCurrentFrame().Locals[1]
				bs := make([]byte, len)
				_, err := r.ReadAt(vm.Memory, bs, int64(ptr))
				if err != nil {
					panic(err)
				}

				val, _ := r.State.databaseApi.GetStorage(string(bs), true)
				if val != nil {
					return 1
				}
				return 0
			}
		case "debug":
			return func(vm *exec.VirtualMachine) int64 {
				if !vm.AddAndCheckGas(DEBUG_GAS) {
					vm.GasLimitExceeded = true
				}
				ptr := vm.GetCurrentFrame().Locals[0]
				len := vm.GetCurrentFrame().Locals[1]
				buf := make([]byte, len)
				_, err := r.ReadAt(vm.Memory, buf, int64(ptr))
				if err != nil {
					panic(err)
				}
				fmt.Println(string(buf))
				return 0
			}
		default:
			panic(fmt.Errorf("unknown field: %s", field))
		}
	default:
		panic(fmt.Errorf("unknown module: %s", module))
	}
}

// ReadAt implements the ReaderAt interface: it copies into p
// the content of memory at offset off.
func (r *Resolver) ReadAt(mem []byte, p []byte, off int64) (int, error) {
	var length int
	if len(mem) < len(p)+int(off) {
		length = len(mem) - int(off)
	} else {
		length = len(p)
	}

	copy(p, mem[off:off+int64(length)])

	var err error
	if length < len(p) {
		err = io.ErrShortBuffer
	}

	return length, err
}

// WriteAt implements the WriterAt interface: it writes the content of p
// into the VM memory at offset off.
func (r *Resolver) WriteAt(mem []byte, p []byte, off int64) (int, error) {
	var length int
	if len(mem) < len(p)+int(off) {
		length = len(mem) - int(off)
	} else {
		length = len(p)
	}

	copy(mem[off:], p[:length])

	var err error
	if length < len(p) {
		err = io.ErrShortWrite
	}

	return length, err
}
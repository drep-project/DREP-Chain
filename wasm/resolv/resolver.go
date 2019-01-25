package resolv

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/perlin-network/life/exec"
)

type Resolver struct {
	Stderr       bytes.Buffer
	BlockedCalls []FCall
	PendingCalls []FCall
	NewCalls     int
}

func New() *Resolver {
	return &Resolver{}
}

func (r *Resolver) Resume(vm *exec.VirtualMachine, c FCall) (int64, error) {
	entryID, ok := vm.GetFunctionExport("run") // can change to whatever exported function name you want
	if !ok {
		return 0, fmt.Errorf("entry function not found")
	}
	for i, v := range r.BlockedCalls {
		if v.CB == c.CB {
			r.Stderr = bytes.Buffer{}
			r.BlockedCalls = append(r.BlockedCalls[:i], r.BlockedCalls[i+1:]...)
			r.PendingCalls = append(r.PendingCalls, c)
			return vm.Run(entryID, 0, 0)
		}
	}
	return 0, fmt.Errorf("Pending call cb %v not found", c.CB)
}

func (r *Resolver) ResolveGlobal(module, field string) int64 {
	return 0
}

// resolve minimal imports required to run Go
func (r *Resolver) ResolveFunc(module, field string) exec.FunctionImport {
	if module == "go" {
		switch field {
		case "runtime.wasmWrite":
			return r.Write
		case "runtime.nanotime":
			return r.NanoTime
		case "runtime.walltime":
			return r.WallTime
		case "syscall/js.valueGet":
			return r.ValueGet
		case "syscall/js.valueCall":
			return r.ValueCall
		case "syscall/js.valueInvoke":
			return r.ValueInvoke
		case "syscall/js.valueNew":
			return r.ValueNew
		case "syscall/js.valueLength":
			return r.ValueLength
		case "syscall/js.valueIndex":
			return r.ValueIndex
		case "syscall/js.valuePrepareString":
			return r.ValuePrepString
		case "syscall/js.valueLoadString":
			return r.ValueLoadString
		case "runtime.getRandomData":
			return r.RandomData
		case "runtime.wasmExit":
			return r.WasmExit
		default:
			return r.Stub(field)
		}
	}
	return nil
}

func (r *Resolver) Stub(name string) func(vm *exec.VirtualMachine) int64 {
	return func(vm *exec.VirtualMachine) int64 {
		panic(name)
		return 0
	}
}

func (r *Resolver) RandomData(vm *exec.VirtualMachine) int64 {
	return 0
}

func (r *Resolver) WasmExit(vm *exec.VirtualMachine) int64 {
	return 0
}

func (r *Resolver) ValueLength(vm *exec.VirtualMachine) int64 {
	setInt64(vm, 16, 1)
	return 0
}
func (r *Resolver) ValueIndex(vm *exec.VirtualMachine) int64 {
	setInt64(vm, 16, 1)
	return 0
}
func (r *Resolver) ValuePrepString(vm *exec.VirtualMachine) int64 {
	setInt64(vm, 16, jsCurCbStr)
	setInt64(vm, 24, int64(len(curCB.Output)))
	return 0
}
func (r *Resolver) ValueLoadString(vm *exec.VirtualMachine) int64 {
	arr := getInt64(vm, 16)
	len := getInt64(vm, 24)
	copy(vm.Memory[arr:arr+len], curCB.Output)
	return 0
}

func (r *Resolver) WallTime(vm *exec.VirtualMachine) int64 {
	setInt64(vm, 8, time.Now().Unix())
	setInt64(vm, 16, time.Now().UnixNano()/1000000000)
	return 0
}

func (r *Resolver) NanoTime(vm *exec.VirtualMachine) int64 {
	setInt64(vm, 8, time.Now().UnixNano())
	return 0
}

func (r *Resolver) Write(vm *exec.VirtualMachine) int64 {
	sp := int(uint32(vm.GetCurrentFrame().Locals[0]))
	fd := binary.LittleEndian.Uint64(vm.Memory[sp+8 : sp+16])
	ptr := binary.LittleEndian.Uint64(vm.Memory[sp+16 : sp+24])
	msgLen := binary.LittleEndian.Uint64(vm.Memory[sp+24 : sp+32])
	msg := vm.Memory[ptr : ptr+msgLen]
	var out io.Writer
	switch fd {
	case 2:
		out = &r.Stderr
	default:
		panic("only stderr file descriptor is supported")
	}
	n, err := out.Write(msg)
	if err != nil {
		panic(err)
	}
	return int64(n)
}

// function parameters as JSON in method name to
// avoid unnecessary conversion logic
type FCall struct {
	Method string
	CB     int
	Input  json.RawMessage
	Output json.RawMessage
}

// this is invoked only for _makeCallbackHelper, which we don't need
// so we just ignore the stuff
func (r *Resolver) ValueInvoke(vm *exec.VirtualMachine) int64 {
	setInt8(vm, 48, 1)
	return 0
}

func (r *Resolver) ValueCall(vm *exec.VirtualMachine) int64 {
	ptr := getInt64(vm, 8)
	b := loadBytes(vm, 16)
	if ptr == jsPendCbs && string(b) == "shift" {
		if len(r.PendingCalls) == 0 {
			setInt64(vm, 56, jsZero)
			setInt8(vm, 64, 1)
			return 0
		}
		curCB, r.PendingCalls = r.PendingCalls[0], r.PendingCalls[1:]
		setInt64(vm, 56, jsCurCb)
		setInt8(vm, 64, 1)
		return 0
	}
	if ptr != jsWiasm {
		panic(fmt.Sprintf("Call implemented only for kwwasm object: %v %v", ptr, string(b)))
	}
	var c FCall
	err := json.Unmarshal(b, &c)
	if err != nil {
		panic(err)
	}
	switch c.Method {
	case "wiasm.log":
		var inp string
		err := json.Unmarshal(c.Input, &inp)
		if err != nil {
			panic(err)
		}
		fmt.Println(inp)
		_, err = r.Stderr.WriteString(inp)
		if err != nil {
			panic(err)
		}
	default:
	    fmt.Println("resov.ValueCall" + c.Method)
		r.BlockedCalls = append(r.BlockedCalls, c)
	}
	setInt8(vm, 64, 1)
	return 0
}

func (r *Resolver) ValueNew(vm *exec.VirtualMachine) int64 {
	if r.NewCalls > 0 {
		panic("only single call to New() is allowed! (for pending Callbacks)")
	}
	r.NewCalls++ // prevent any use of syscall/js API, i.e. fail fast
	setInt64(vm, 40, jsPendCbs)
	setInt64(vm, 48, 1)
	return 0
}

// because VM should not have business with host's memory
// we have a set of hardcoded pointers & values that always have
// consistent value across all VM hosts.
// We don't wanna manage garbage collection and other shit
// the whole point is to make minimal functionality to run GO inside life VM
func (r *Resolver) ValueGet(vm *exec.VirtualMachine) int64 {
	ptr := getInt64(vm, 8)
	str := loadString(vm, 16)
	prefix := "?."
	switch ptr {
	case jsGlobal:
		prefix = "jsGlobal."
	case linMem:
		prefix = "linMem."
	case jsMemBuf:
		prefix = "jsMemBuf."
	case jsStub:
		prefix = "jsStub."
	case jsZero:
		prefix = "jsZero."
	case jsGo:
		prefix = "jsGo."
	case jsWiasm:
		prefix = "jsWiasm."
	case jsFs:
		prefix = "jsFs."
	case jsProcess:
		prefix = "jsProcess."
	case jsCurCb:
		prefix = "jsCurCb."
	}
	name := prefix + str
	switch name {
	case "jsGlobal.Go":
		setInt64(vm, 32, jsGo)
		return 0
	case "jsGlobal.process":
		setInt64(vm, 32, jsProcess)
		return 0
	case "linMem.buffer":
		setInt64(vm, 32, jsMemBuf)
		return 0
	case "jsGlobal.fs":
		setInt64(vm, 32, jsFs)
		return 0
	case "jsGo._callbackShutdown":
		setInt64(vm, 32, jsFalse)
		return 0
	case "jsGlobal.wiasm":
		setInt64(vm, 32, jsWiasm)
		return 0
	case "jsGo._makeCallbackHelper":
		setInt64(vm, 32, cbHelper)
		return 0
	case "jsFs.O_WRONLY":
		setJsInt(vm, 32, -1)
		return 0
	case "jsFs.O_RDWR":
		setJsInt(vm, 32, -1)
		return 0
	case "jsFs.O_CREAT":
		setJsInt(vm, 32, -1)
		return 0
	case "jsFs.O_TRUNC":
		setJsInt(vm, 32, -1)
		return 0
	case "jsFs.O_APPEND":
		setJsInt(vm, 32, -1)
		return 0
	case "jsFs.O_EXCL":
		setJsInt(vm, 32, -1)
		return 0
	case "jsFs.constants":
		return 0 // stub
	case "jsCurCb.id":
		setJsInt(vm, 32, int64(curCB.CB))
		return 0
	case "jsCurCb.args":
		setJsInt(vm, 32, jsCurCbArgs)
		return 0
	}
	setInt64(vm, 32, jsStub)
	return 0
}

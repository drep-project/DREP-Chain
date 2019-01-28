package wasm

import (
    "io/ioutil"
    "github.com/perlin-network/life/exec"
    "rep_algorithm/resolv"
    "encoding/json"
    "fmt"
    "log"
)

type Function struct  {
    MethodName string
    Params interface{}
    Result interface{}
}

func readWasm() []byte {
    b, err := ioutil.ReadFile("wasm/app.wasm")
    if err != nil {
        log.Fatal("file read:", err)
    }
    return b
}

func setupVmAndResolv(b []byte) (r *resolv.Resolver, vm *exec.VirtualMachine) {
    r = resolv.New()
    vm, err := exec.NewVirtualMachine(b, exec.VMConfig{}, r, nil)
    if err != nil { // if the wasm bytecode is invalid
        log.Fatal("vm create:", err)
    }
    entryID, ok := vm.GetFunctionExport("run") // can change to whatever exported function name you want
    if !ok {
        panic("entry function not found")
    }
    _, err = vm.Run(entryID, 0, 0) // start vm
    if err != nil {
        vm.PrintStackTrace()
        log.Fatal("vm run:", err)
    }
    return r, vm
}

func generateInput(f Function) []byte {
    //Input := make(map[string] interface{})
    //Input["method"] = f.MethodName
    //
    //Input["params"] = f.Params
    //bytes, err:= json.Marshal(Input)
    bytes, err:= json.Marshal(f)
    if err != nil {
        fmt.Println(err.Error())
    }
    return bytes
}


func resumeCallFunc(vm *exec.VirtualMachine, r *resolv.Resolver, input []byte, index int)  {
    ret, err := r.Resume(vm, resolv.FCall{ // resume vm execution with callback result
        CB:     r.BlockedCalls[index].CB,
        Output: input,
    })

    if err != nil {
        vm.PrintStackTrace()
        log.Fatal("1111 vm run:", err)
    }
    log.Printf("ret: %v, log:%v", ret, r.Stderr.String())
}

func resumeReturnCall(vm *exec.VirtualMachine, r *resolv.Resolver, methodName string) string {
    for index, v := range r.BlockedCalls {
        fmt.Println("resumeReturnCall:" + v.Method)
        if v.Method == methodName {
            result := string(r.BlockedCalls[index].Input)
            input, _ := json.Marshal(make(map[string] string))
            resumeCallFunc(vm, r, input, index)
            return result
        }
    }
    return ""
}

func callFunc(vm *exec.VirtualMachine, r *resolv.Resolver, f Function) string {
    if len(r.BlockedCalls) > 0 {
        isExist := false
        for index, v := range r.BlockedCalls {
            fmt.Println("callFunc:" + v.Method)
            if v.Method == f.MethodName {
                isExist = true
                input := generateInput(f)
                fmt.Println("executing :"  + string(input) )
                resumeCallFunc(vm, r, input, index)
            }
        }

        return resumeReturnCall(vm, r, f.MethodName + "Result")

        if !isExist {
            fmt.Println("there's no function named " + f.MethodName + " exported in wasm file!")
        }
    }
    return ""
}

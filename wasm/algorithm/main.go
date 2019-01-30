package main

import "BlockChainTest/wasm/wiasm"

var methodsMap = map[string] WasmFunc {
    "RegisterUserByParams": RegisterUserByParams,
    "GainByParams":         GainByParams,
    "LiquidateByParams":    LiquidateByParams,
    "AcceptModel":          AcceptModel,
}

func main() {
    channel := make(chan *Function)
    go func() {
        call := &Function{"RegisterUserByParams","",""}
        wiasm.Call("RegisterUserByParams", "", &call)
        channel <- call
    }()
    go func() {
        call := &Function{"GainByParams","",""}
        wiasm.Call(call.MethodName, "", &call)
        channel <- call
    }()
    go func() {
        call := &Function{"AcceptModel","",""}
        wiasm.Call(call.MethodName, "", &call)
        channel <- call
    }()
    go func() {
        call := &Function{"LiquidateByParams","",""}
        wiasm.Call(call.MethodName, "", &call)
        channel <- call
    }()
    for i := 0; i < 100; i++ {
        select {
        case call := <-channel:
            exectueDecision(channel, call)
        }
    }
}

func resultCallBack(channel chan *Function, method string, result interface{})  {
    c := make(chan *Function)
    go func() {
        var callback *Function
        wiasm.Call(method + "Result", result, &callback)
        c <- callback

        var call *Function
        wiasm.Call(method, "", &call)
        channel <- call
        select {
        case call := <-channel:
            exectueDecision(channel, call)
        }
    }()
    select {
    case <- c:
        wiasm.Log("result return finished ")
    }
}

func exectueDecision(channel chan *Function, call *Function)  {
    for pattern, handleFunc := range (methodsMap) {
        if pattern == call.MethodName {
            resultCallBack(channel, pattern, handleFunc(call.Params))
        }
    }
}

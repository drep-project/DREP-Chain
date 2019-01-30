package wiasm

// +build js,wasm

import (
	"encoding/json"
	"syscall/js"
	"unsafe"
	"BlockChainTest/log"
)

type FCall struct {
	Method string
	CB     int
	Input  json.RawMessage
}

func Exit() {
	Call("wiasm.exit", nil, nil)
}

func Log(str string) {
	in, _ := json.Marshal(str + "\n")
	b, _ := json.Marshal(FCall{
		Method: "wiasm.log",
		Input:  in,
	})
	js.Global().Get("wiasm").Call(string(b))
}

type jsCB struct {
	ref int64
	id  uint32
}

func Call(method string, input interface{}, output interface{}) {
	done := make(chan string)
	defer close(done)
	cb := js.NewCallback(func(args []js.Value) {
		done <- args[0].String()
	})
	u := *((*jsCB)(unsafe.Pointer(&cb)))
	in, _ := json.Marshal(input)
	b, _ := json.Marshal(FCall{
		Method: method,
		CB:     int(u.id),
		Input:  in,
	})
	js.Global().Get("wiasm").Call(string(b))
	data := <-done // this will unblock when pendingCallbacc is set and VM resumed
	log.Printf("51line log.Printf: " + data)
	err := json.Unmarshal([]byte(data), output)
	if err != nil {
		panic(err)
	}
}

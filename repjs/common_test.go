package repjs

import (
    "testing"
    "encoding/json"
    "math/big"
    "fmt"
)

func initTracer() map[string] interface{} {
    tracer := make(map[string] interface{})
    tracer["Rep"], _ = new(big.Float).SetString("12.22")
    tracer["Recent"], _ = new(big.Float).SetString("6.66")
    tracer["Remote"], _ = new(big.Float).SetString("7.77")
    tracer["Aliveness"], _ = new(big.Float).SetString("22")
    tracer["Bottom"], _ = new(big.Float).SetString("10.01")
    tracer["GainMemory"] = make(map[int] interface{})
    tracer["GainMemory"].(map[int] interface{})[1] = 10;
    tracer["GainMemory"].(map[int] interface{})[2] = 20;
    tracer["GainHistory"] = make([]int, 30)
    tracer["GainHistory"].([]int)[1] = 10
    tracer["GainHistory"].([]int)[2] = 20
    tracer["FirstActiveDay"] = 1
    tracer["LastLiquidateDay"] = 0
    return tracer
}

func initActive() bool {
    active := false
    return active
}

func initGroup() []string {
    group := make([]string, 10)
    group[0] = "ae32acf43"
    group[1] = "4342nc7ee"
    group[2] = "bb977ac21"
    return group
}

func TestBytes2TracerValue(t *testing.T) {
    tracer := initTracer()
    b, _ := json.Marshal(tracer)
    fmt.Println("result: ", bytes2TracerValue(b))
}

func TestTracerValue2Bytes(t *testing.T) {
    tracer := initTracer()
    b, _ := json.Marshal(tracer)
    result := bytes2TracerValue(b)
    fmt.Println("js value: ", tracerValue2Bytes(result))
}

func TestBytes2ActiveValue(t *testing.T) {
    active := initActive()
    b, _ := json.Marshal(active)
    fmt.Println("result: ", bytes2ActiveValue(b))
}

func TestActiveValue2Bytes(t *testing.T) {
    active := initActive()
    b, _ := json.Marshal(active)
    result := bytes2ActiveValue(b)
    fmt.Println("js value: ", activeValue2Bytes(result))
}

func TestBytes2GroupValue(t *testing.T) {
    group := initGroup()
    b, _ := json.Marshal(group)
    fmt.Println("result: ", bytes2GroupValue(b))
}

func TestGroupValue2Bytes(t *testing.T) {
    group := initGroup()
    b, _ := json.Marshal(group)
    result := bytes2GroupValue(b)
    fmt.Println("js value: ", groupValue2Bytes(result))
}
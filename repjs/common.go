package repjs

import (
    "github.com/robertkrimen/otto"
    "strconv"
    "encoding/json"
)

func bytes2TracerValue(b []byte) otto.Value {
    if b == nil {
        return otto.UndefinedValue()
    }
    jObject, err := vm.Object("(" + string(b) + ")")
    if err != nil {
        return otto.UndefinedValue()
    }
    return jObject.Value()
}

func tracerValue2Bytes(jValue otto.Value) []byte {
    tracer, err := jValue.Export()
    if err != nil {
        return nil
    }
    b, err := json.Marshal(tracer)
    if err != nil {
        return nil
    }
    return b
}

func bytes2ActiveValue(b []byte) otto.Value {
    if b == nil {
        return otto.UndefinedValue()
    }
    ok, err := strconv.ParseBool(string(b))
    if err != nil {
        return otto.UndefinedValue()
    }
    jValue, err := vm.ToValue(ok)
    if err != nil {
        return otto.UndefinedValue()
    }
    return jValue
}

func activeValue2Bytes(jValue otto.Value) []byte {
    active, err := jValue.ToBoolean()
    if err != nil {
        return nil
    }
    b := []byte(strconv.FormatBool(active))
    return b
}

func bytes2GroupValue(b []byte) otto.Value {
    if b == nil {
        return otto.UndefinedValue()
    }
    jObject, err := vm.Object(string(b))
    if err != nil {
        return otto.UndefinedValue()
    }
    return jObject.Value()
}

func groupValue2Bytes(jValue otto.Value) []byte {
    repIds, err  := jValue.Export()
    if err != nil {
        return nil
    }
    b, err := json.Marshal(repIds)
    if err != nil {
        return nil
    }
    return b
}
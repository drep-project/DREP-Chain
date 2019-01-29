package repjs

import (
    "github.com/robertkrimen/otto"
    "fmt"
    "strings"
    "os"
    "path"
    "io/ioutil"
    "errors"
    "math/big"
    "BlockChainTest/mycrypto"
    "BlockChainTest/database"
    "encoding/hex"
)

var (
    vm = otto.New()
    groupNum = 5
)

func init() {
    vm.Set("require", require)

    consoleObj, _ := vm.Get("console")
    consoleObj.Object().Set("log", consoleOutput)

    cryptoObject,_ := vm.Object(`({ })`)
    vm.Set("crypto", cryptoObject)
    cryptoObject.Set("hash256", hash256)

    utilObject,_ := vm.Object(`({ })`)
    vm.Set("utils", utilObject)
    utilObject.Set("str2Bytes", str2Bytes)
    utilObject.Set("bytes2Str", bytes2Str)
    utilObject.Set("bytes2Hex", bytes2Hex)
    utilObject.Set("allocateGroupID", allocateGroupID)

    dbObject,_ := vm.Object(`({ })`)
    vm.Set("db", dbObject)
    dbObject.Set("getTracer", getTracer)
    dbObject.Set("putTracer", putTracer)
    dbObject.Set("setActive", setActive)
    dbObject.Set("isActive", isActive)
    dbObject.Set("getGroup", getGroup)
    dbObject.Set("putGroup", putGroup)

    repObject, _ := vm.Object(`({ })`)
    vm.Set("rep", repObject)

    _, err := vm.Run("var reputation = require('repjs/reputation.js')")
    if err != nil {
        fmt.Println(err.Error())
    }
}

func str2Bytes(str string) []byte {
    return []byte(str)
}

func bytes2Str(b []byte) string {
    return string(b)
}

func bytes2Hex(b []byte) string {
    return hex.EncodeToString(b)
}

func hash256(data ...[]byte) []byte {
    return mycrypto.Hash256(data...)
}

func allocateGroupID(repID string) uint64 {
   i, _ := new(big.Int).SetString(repID, 16)
   num := new(big.Int).SetInt64(int64(groupNum - 2))
   one := new(big.Int).SetInt64(1)
   index := i.Mod(i, num)
   index.Add(index, one)
   groupID := uint64(index.Int64())
   return groupID
}

func consoleOutput(call otto.FunctionCall) otto.Value {
    output := make([]string, 0)
    for _, argument := range call.ArgumentList {
        output = append(output, fmt.Sprintf("%v", argument))
    }
    fmt.Println(strings.Join(output, " "))
    return otto.Value{}
}

func getTracer(platformID, repId string) otto.Value {
    return bytes2TracerValue(database.GetTracer(platformID, repId))
}

func putTracer(platformID, repID string, jValue otto.Value) error {
    return database.PutTracer(platformID, repID, tracerValue2Bytes(jValue))
}

func isActive(platformID, repID string) otto.Value {
    return bytes2ActiveValue(database.IsActive(platformID, repID))
}

func setActive(platformID, repID string, jValue otto.Value) error {
    return database.SetActive(platformID, repID, activeValue2Bytes(jValue))
}

func getGroup(platformID string, groupID uint64) otto.Value {
    return bytes2GroupValue(database.GetGroup(platformID, groupID))
}

func putGroup(platformID string, groupID uint64, jValue otto.Value) error {
    return database.PutGroup(platformID, groupID, groupValue2Bytes(jValue))
}

func require(call otto.FunctionCall) otto.Value {
    file := call.Argument(0).String()
    fmt.Printf("requiring: %s\n", file)
    fi, err := os.Stat(file)
    if err != nil {
        fmt.Println(err)
        panic(err)
    }
    fileSuffix := path.Ext(fi.Name())

    data, err := ioutil.ReadFile(file)
    if err != nil {
        fmt.Println(err)
        panic(err)
    }
    var jVal otto.Value
    if fileSuffix == ".json" {
        jVal, err = call.Otto.Run("(" + string(data) + ")")
    } else  if fileSuffix == ".js"{
        jVal, err = call.Otto.Run(string(data))
    }else{
        panic(errors.New("unsupport file type:"+ fi.Name()))
    }
    if err != nil {
        fmt.Println(err)
        panic(err)
    }
    return jVal
}

func GetProfile(platformID, uid string) (map[string] interface{}) {
    fun, err := vm.Get("getProfile")
    if err != nil {
        return nil
    }
    ret, err := fun.Call(vm.Context().This, platformID, uid)
    if err != nil {
        return nil
    }
    data, err := ret.Export()
    if err != nil {
        return nil
    }
    if value, ok := data.(map[string] interface{}); ok {
        return value
    }
    return nil
}

func RegisterUser(platformID, repID string, groupID uint64) error {
    fun, err := vm.Get("registerUser")
    if err != nil {
        return err
    }
    _, err = fun.Call(vm.Context().This, platformID, repID, groupID)
    return err
}

//func RegisterUsers(platformID, UIDs []string) ([]map[string] string, error) {
//    fun, err := vm.Get("registerUsers")
//    if err != nil {
//        return nil, err
//    }
//    ret, err := fun.Call(vm.Context().This, platformID, UIDs)
//    if err != nil {
//        return nil, err
//    }
//    data, err := ret.Export()
//    if err != nil {
//        return nil, err
//    }
//    if value, ok := data.([]map[string] string); ok {
//        return value, nil
//    }
//    return nil, errors.New("wrong type returns")
//}

func AddGain(platformID string, increments []map[string] interface{}) error {
    fun, err := vm.Get("addGain")
    if err != nil {
        return err
    }
    _, err = fun.Call(vm.Context().This, platformID, increments)
    return err
}

//func LiquidateRep(platformID string, repIDs []string, until int) error {
//    fun, err := vm.Get("liquidateRep")
//    if err != nil {
//        return err
//    }
//    _, err = fun.Call(vm.Context().This, platformID, repIDs, until)
//    return err
//}

func LiquidateRepByGroup(platformID string, groupID uint64, until int) (map[string] interface{}, error) {
    fun, err := vm.Get("liquidateRepByGroup")
    if err != nil {
        return nil, err
    }
    ret, err := fun.Call(vm.Context().This, platformID, groupID, until)
    if err != nil {
        return nil, err
    }
    fmt.Println("ret: ", ret)
    data, err := ret.Export()
    if err != nil {
        return nil, err
    }
    fmt.Println("data: ", data)
    if value, ok := data.(map[string] interface{}); ok {
        return value, nil
    }
    return nil, errors.New("wrong js value type")
}

func LiquidateRepByGroupSimply(platformID string, groupID uint64, until int) (map[string] interface{}, error) {
    fun, err := vm.Get("liquidateRepByGroupSimply")
    if err != nil {
        return nil, err
    }
    ret, err := fun.Call(vm.Context().This, platformID, groupID, until)
    if err != nil {
        return nil, err
    }
    fmt.Println("ret: ", ret)
    data, err := ret.Export()
    if err != nil {
        return nil, err
    }
    fmt.Println("data: ", data)
    if value, ok := data.(map[string] interface{}); ok {
        return value, nil
    }
    return nil, errors.New("wrong js value type")
}
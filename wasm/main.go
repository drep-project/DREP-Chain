package wasm

import (
    "fmt"
    "time"
    "strconv"
    "rep_algorithm/resolv"
    "github.com/perlin-network/life/exec"
    "encoding/json"
    "BlockChainTest/database"
)

var n = 200

func main() {
    b := readWasm()
    r, vm := setupVmAndResolv(b)

    setupModel(vm, r)
    reg_resp := RegisterUser(vm, r)

    users := [] registerReturns{}
    err := json.Unmarshal([]byte(reg_resp), &users)
    if err != nil {
        fmt.Println("json ummarshal users error")
    }

    AddGain(vm, r, users)
    Liquidate(vm, r, users)
}

func setupModel(vm *exec.VirtualMachine, r *resolv.Resolver)  {
    time1 := time.Now()
    params := generateAcceptModelParams()
    err := callFunc(vm, r, Function{"AcceptModel",params,""})
    fmt.Println("AcceptModel error:", err)
    fmt.Println("AcceptModel time:", time.Now().Sub(time1))
}

func RegisterUser(vm *exec.VirtualMachine, r *resolv.Resolver) string {
    params := []string{}
    for i := 0; i < n; i++  {
        uid := database.UID("user_" + strconv.Itoa(i))
        p := generateRegisterParams("a", uid)
        params = append(params, p)
    }

    time1 := time.Now()
    resp := callFunc(vm, r, Function{"RegisterUserByParams",params,""})
    fmt.Println("RegisterUser time:", time.Now().Sub(time1))
    fmt.Println("RegisterUser result: ", resp)
    return resp
}

func AddGain(vm *exec.VirtualMachine, r *resolv.Resolver, users []registerReturns)  {
    time1 := time.Now()
    increments := []*gainIncrement{}
    for _, user := range users {
        id := processRegisterReturns(&user)
        increment := &gainIncrement{id, 30, 1}
        increments = append(increments, increment)
    }

    p := generateGainParams("a", increments)
    resp := callFunc(vm, r,  Function{"GainByParams",p,""})

    fmt.Println("AddGain resp: ", resp)
    fmt.Println("AddGain time:", time.Now().Sub(time1))
    processGainReturns(resp)
}

func Liquidate(vm *exec.VirtualMachine, r *resolv.Resolver, users []registerReturns)  {
    time1 := time.Now()
    ids := []database.RepID{}
    for _, user := range users {
        ids = append(ids, user.RepID)
    }
    params := generateLiquidateParams("a", 2, ids)
    fmt.Println("Liquidate params: ", params)
    resp := callFunc(vm, r, Function{"LiquidateByParams",params,""})
    fmt.Println("Liquidate resp: ", resp)
    fmt.Println("Liquidate time:", time.Now().Sub(time1))
}

func demo() {
    //time1 := time.Now()
    //params := generateAcceptModelParams()
    //err := callFunc(vm, r, Function{"AcceptModel",params,""})
    //fmt.Println("AcceptModel error:", err)
    //fmt.Println("AcceptModel time:", time.Now().Sub(time1))
    //
    //time1 = time.Now()
    //p1 := generateRegisterParams("a", "eric")
    //r1 := callFunc(vm, r, Function{"RegisterUserByParams",p1,""})
    //fmt.Println("result1:" + r1)
    //fmt.Println("callRegisterUser time:", time.Now().Sub(time1))
    //
    //id1, _ := processRegisterReturns(r1)
    //fmt.Println("id1: ", id1)
    //
    //p5 := generateGainParams("a", []*gainIncrement{{id1, 20, 1}})
    //
    //time1 = time.Now()
    //fmt.Println(time1)
    //r2 := callFunc(vm, r,  Function{"GainByParams",p5,""})
    //fmt.Println("result2:" + r2)
    //fmt.Println("callGainByParams time:", time.Now().Sub(time1))
    //time1 = time.Now()
    //fmt.Println(time1)
    //processGainReturns(r2)
    //
    //fmt.Println(getTracer("a", id1))
    //p6 := generateLiquidateParams("a", 2, []RepID{id1})
    //fmt.Println("p6: ", p6)
    //fmt.Println("param_acpt: ", mod)
    //
    //time1 = time.Now()
    //r6 := callFunc(vm, r, Function{"LiquidateByParams",p6,""})
    //fmt.Println("r6: ", r6)
    //time2 = time.Now()
    //fmt.Println("callLiquidateByParams time:", time2.Sub(time1))
}
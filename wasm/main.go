package wasm

import (
    "fmt"
    "time"
    "rep_algorithm/resolv"
    "github.com/perlin-network/life/exec"
    "encoding/json"
    "BlockChainTest/database"
    "strconv"
)

var (
    n = 100
    vm *exec.VirtualMachine
    r *resolv.Resolver
    isRegisterUser = false
    ids = []database.RepID{}
)

func init()  {
    b := readWasm()
    r, vm = setupVmAndResolv(b)
    setupModel()
}

func main() {
    var uids []string
    for i := 0; i < n; i++  {
        uid := "user_" + strconv.Itoa(i)
        uids = append(uids, uid)
    }
    reg_resp := RegisterUser(uids)

    users := [] RegisterReturns{}
    err := json.Unmarshal([]byte(reg_resp), &users)
    if err != nil {
        fmt.Println("json ummarshal users error")
    }

    AddGain(users, 1)
    Liquidate(users, 2)
}

func setupModel()  {
    time1 := time.Now()
    params := generateAcceptModelParams()
    err := callFunc(vm, r, Function{"AcceptModel",params,""})
    fmt.Println("AcceptModel error:", err)
    fmt.Println("AcceptModel time:", time.Now().Sub(time1))
}

func RegisterUser(uids []string) string {
    params := []string{}

    for _, uid := range uids {
        p := generateRegisterParams("a", database.UID(uid))
        params = append(params, p)
    }

    time1 := time.Now()
    resp := callFunc(vm, r, Function{"RegisterUserByParams",params,""})
    fmt.Println("RegisterUser time:", time.Now().Sub(time1))
    fmt.Println("RegisterUser result: ", resp)
    return resp
}

func AddGain(users []RegisterReturns, height int64)  {
    time1 := time.Now()
    increments := []*gainIncrement{}

    if !isRegisterUser {
        writeUsersToDatabase(users)
        isRegisterUser = true
    }

    for _, id := range ids {
        increment := &gainIncrement{id, 30, height}
        increments = append(increments, increment)
    }

    p := generateGainParams("a", increments)
    resp := callFunc(vm, r,  Function{"GainByParams",p,""})

    fmt.Println("AddGain resp: ", resp)
    fmt.Println("AddGain time:", time.Now().Sub(time1))
    processGainReturns(resp)
}

func writeUsersToDatabase(users []RegisterReturns) {
    fmt.Println("wasm writeUsersToDatabase")
    for _, user := range users {
        id := processRegisterReturns(&user)
        ids = append(ids, id)
    }
}

func Liquidate(users []RegisterReturns, height int64)  {
    time1 := time.Now()
    ids := []database.RepID{}
    for _, user := range users {
        ids = append(ids, user.RepID)
    }
    params := generateLiquidateParams("a", height, ids)
    fmt.Println("Liquidate params: ", params)
    resp := callFunc(vm, r, Function{"LiquidateByParams",params,""})
    fmt.Println("Liquidate resp: ", resp)
    fmt.Println("Liquidate time:", time.Now().Sub(time1))
    processLiquidateReturns(resp)
}
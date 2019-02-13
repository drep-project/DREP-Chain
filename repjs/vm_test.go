package repjs

import (
    "fmt"
    "testing"
)

var (
    platformID = "sample_platform"
    eric, sai, long, xie map[string] interface{}
    ericID, saiID, longID, xieID string
    ericGP, saiGP, longGP, xieGP uint64
)

func initAccount() {
    eric = GetProfile(platformID, "eric")
    sai = GetProfile(platformID, "sai")
    long = GetProfile(platformID, "long")
    xie = GetProfile(platformID, "xie")
    ericID = eric["RepID"].(string)
    ericGP = eric["GroupID"].(uint64)
    saiID = sai["RepID"].(string)
    saiGP = sai["GroupID"].(uint64)
    longID = long["RepID"].(string)
    longGP = long["GroupID"].(uint64)
    xieID = xie["RepID"].(string)
    xieGP = xie["GroupID"].(uint64)
    fmt.Println("eric:")
    fmt.Println(ericID, ericGP)
    fmt.Println("sai:")
    fmt.Println(saiID, saiGP)
    fmt.Println("long:")
    fmt.Println(longID, longGP)
    fmt.Println("xie:")
    fmt.Println(xieID, xieGP)
}

func TestRegisterUser(t *testing.T) {
    initAccount()
    RegisterUser(platformID, ericID, ericGP)
    RegisterUser(platformID, saiID, saiGP)
    RegisterUser(platformID, longID, longGP)
    RegisterUser(platformID, xieID, xieGP)
}
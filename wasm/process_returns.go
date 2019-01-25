package wasm

import (
    "encoding/json"
    "errors"
    "BlockChainTest/database"
)

type registerReturns struct {
    PlatformID database.PlatformID
    RepID      database.RepID
    GroupID    database.GroupID
    Tracer     *database.RepTracer
    Error      string
}

type gainReturns struct {
    PlatformID database.PlatformID
    Tracers    map[database.RepID] *database.RepTracer
    Active     map[database.RepID] bool
    Error      string
}

type liquidateReturns struct {
    PlatformID database.PlatformID
    Tracers    map[database.RepID] *database.RepTracer
    Active     map[database.RepID] bool
    Error      string
}

func processRegisterReturns(ret *registerReturns) database.RepID {
    group := database.GetGroupW(ret.PlatformID, ret.GroupID)
    group = append(group, ret.RepID)
    database.SetGroupW(ret.PlatformID, ret.GroupID, group)
    database.SetActiveW(ret.PlatformID, ret.RepID, false)
    database.PutTracerW(ret.PlatformID, ret.RepID, ret.Tracer)
    return ret.RepID
}

func processGainReturns(returns string) error {
    ret := &gainReturns{}
    err := json.Unmarshal([]byte(returns), ret)
    if err != nil {
        return err
    }
    if ret.Error != "" {
        return errors.New(ret.Error)
    }
    for repID, tracer := range ret.Tracers {
        database.PutTracerW(ret.PlatformID, repID, tracer)
    }
    for repID, active := range ret.Active {
        database.SetActiveW(ret.PlatformID, repID, active)
    }
    return nil
}

func processLiquidateReturns(returns string) error {
    ret := &liquidateReturns{}
    err := json.Unmarshal([]byte(returns), ret)
    if err != nil {
        return err
    }
    if ret.Error != "" {
        return errors.New(ret.Error)
    }
    for repID, tracer := range ret.Tracers {
        database.PutTracerW(ret.PlatformID, repID, tracer)
    }
    for repID, active := range ret.Active {
        database.SetActiveW(ret.PlatformID, repID, active)
    }
    return nil
}
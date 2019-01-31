package wasm

import (
    "math/big"
    "encoding/json"
    "BlockChainTest/database"
)

type registerParams struct {
    UID        database.UID
    PlatformID database.PlatformID
}

type gainIncrement struct {
    RepID database.RepID
    Gain  int
    Day   int64
}

type gainParams struct {
    PlatformID database.PlatformID
    Increments []*gainIncrement
    Tracers    map[database.RepID] *database.RepTracer
    Active     map[database.RepID] bool
}

type model struct {
    R0          *big.Float
    T0          *big.Float
    Te          *big.Float
    Alpha1      *big.Float
    Alpha2      map[int] *big.Float
    CutOff      *big.Float
    BottomRate  *big.Float
    Beta        *big.Float
    Epsilon     * big.Float
    Require     *big.Float
}

type liquidateParams struct {
    PlatformID database.PlatformID
    Until      int64
    RepIDs     []database.RepID
    Tracers    map[database.RepID] *database.RepTracer
    Active     map[database.RepID] bool
}

func generateRegisterParams(platformID database.PlatformID, uid database.UID) string {
    par := &registerParams{
        PlatformID: platformID,
        UID:       uid,
    }
    b, err := json.Marshal(par)
    if err != nil {
        return "json marshal params error"
    }
    return string(b)
}

func generateGainParams(platformID database.PlatformID, increments []*gainIncrement) string {
    par := &gainParams{}
    par.PlatformID = platformID
    par.Increments = increments
    par.Tracers = make(map[database.RepID] *database.RepTracer)
    par.Active = make(map[database.RepID] bool)
    for _, inc := range increments {
        if _, ok := par.Tracers[inc.RepID]; !ok {
            tracer := database.GetTracerW(platformID, inc.RepID)
            if tracer != nil {
                par.Tracers[inc.RepID] = tracer
            }
        }
        if _, ok := par.Active[inc.RepID]; !ok {
            par.Active[inc.RepID] = database.IsActiveW(platformID, inc.RepID)
        }
    }
    b, err := json.Marshal(par)
    if err != nil {
        return "json marshal params error"
    }
    return string(b)
}

func generateAcceptModelParams() string{
    mod := SetupModel()
    b, err := json.Marshal(mod)
    if err != nil {
        return "json marshal params error"
    }
    return string(b)
}

func generateLiquidateParams(platformID database.PlatformID, until int64, repIDs []database.RepID) string {
    par := &liquidateParams{}
    par.PlatformID = platformID
    par.RepIDs = repIDs
    par.Until = until
    par.RepIDs = repIDs
    par.Tracers = make(map[database.RepID] *database.RepTracer)
    par.Active = make(map[database.RepID] bool)
    for _, repID := range repIDs {
        if _, ok := par.Tracers[repID]; !ok {
            tracer := database.GetTracerW(platformID, repID)
            if tracer != nil {
                par.Tracers[repID] = tracer
            }
        }

        if _, ok := par.Active[repID]; !ok {
            par.Active[repID] = database.IsActiveW(platformID, repID)
        }
    }
    b, err := json.Marshal(par)
    if err != nil {
        return "json marshal params error"
    }
    return string(b)
}
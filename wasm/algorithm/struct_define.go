package main

import "math/big"

var (
    one      = new(big.Float).SetInt64(int64(1))
    groupNum = 10000
    mod  *model
)

type UID        string
type PlatformID string
type RepID      string
type GroupID    uint64

type WasmFunc func(params interface{}) interface{}

type Function struct  {
    MethodName string
    Params interface{}
    Result interface{}
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

type RepTracer struct {
    Rep              *big.Float
    Aliveness        *big.Float
    Recent           *big.Float
    Remote           *big.Float
    Bottom           *big.Float
    GainHistory      []int
    GainMemory       map[int] int
    FirstActiveDay   int
    LastLiquidateDay int
}

type registerParams struct {
    UID        UID
    PlatformID PlatformID
}

type registerReturns struct {
    PlatformID PlatformID
    RepID      RepID
    GroupID    GroupID
    Tracer     *RepTracer
    Error      string
}

type gainIncrement struct {
    RepID RepID
    Gain  int
    Day   int
}

type gainParams struct {
    PlatformID PlatformID
    Increments []*gainIncrement
    Tracers    map[RepID] *RepTracer
    Active     map[RepID] bool
}

type gainReturns struct {
    PlatformID PlatformID
    Tracers    map[RepID] *RepTracer
    Active     map[RepID] bool
    Error      string
}

type liquidateParams struct {
    PlatformID PlatformID
    Until      int
    RepIDs     []RepID
    Tracers    map[RepID] *RepTracer
    Active     map[RepID] bool
}

type liquidateReturns struct {
    PlatformID PlatformID
    Tracers    map[RepID] *RepTracer
    Active     map[RepID] bool
    Error      string
}

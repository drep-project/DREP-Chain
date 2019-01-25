package main

import (
    "math/big"
    "crypto/sha256"
    "encoding/hex"
)

func hash(b []byte) []byte {
    h := sha256.New()
    h.Write(b)
    return h.Sum(nil)
}

func allocateGroupID(repID RepID) GroupID {
    i, _ := new(big.Int).SetString(string(repID), 16)
    num := new(big.Int).SetInt64(int64(groupNum - 2))
    one := new(big.Int).SetInt64(1)
    index := i.Mod(i, num)
    index.Add(index, one)
    groupID := GroupID(uint64(index.Int64()))
    return groupID
}

func generateOneUser (p *registerParams) *registerReturns {
    user := &registerReturns{}
    b1 := hash([]byte(p.UID))
    b2 := hash([]byte(p.PlatformID))
    b1 = append(b1, b2...)
    b := hash(b1)
    repID := RepID(hex.EncodeToString(b))
    groupID := allocateGroupID(repID)

    tracer := &RepTracer{
        Rep:              new(big.Float),
        Aliveness:        new(big.Float),
        Recent:           new(big.Float),
        Remote:           new(big.Float),
        Bottom:           new(big.Float),
        GainHistory:      make([]int, 0),
        GainMemory:       make(map[int] int),
        FirstActiveDay:   0,
        LastLiquidateDay: 0,
    }
    user.PlatformID = p.PlatformID
    user.RepID = repID
    user.GroupID = groupID
    user.Tracer = tracer
    return user
}


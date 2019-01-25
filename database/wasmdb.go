package database

import (
    "BlockChainTest/mycrypto"
    "strconv"
    "encoding/json"
    "github.com/syndtr/goleveldb/leveldb"
    "math/big"
)

var wasm_db, _ = leveldb.OpenFile("rep_db", nil)

type UID        string
type PlatformID string
type RepID      string
type GroupID    uint64

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

func GetTracerW(platformID PlatformID, repID RepID) *RepTracer {
    key := mycrypto.Hash256([]byte("tracer"), []byte(platformID), []byte(repID))
    value, err := wasm_db.Get(key, nil)
    if err != nil {
        return nil
    }
    tracer := &RepTracer{}
    err = json.Unmarshal(value, tracer)
    if err != nil {
        return nil
    }
    return tracer
}

func PutTracerW(platformID PlatformID, repID RepID, tracer *RepTracer) error {
    key := mycrypto.Hash256([]byte("tracer"), []byte(platformID), []byte(repID))
    value, err := json.Marshal(tracer)
    if err != nil {
        return err
    }
    return wasm_db.Put(key, value, nil)
}

func IsActiveW(platformID PlatformID, repID RepID) bool {
    key := mycrypto.Hash256([]byte("active_state"), []byte(platformID), []byte(repID))
    value, err := wasm_db.Get(key, nil)
    if err != nil {
        return false
    }
    b, err := strconv.ParseBool(string(value))
    if err != nil {
        return false
    }
    return b
}

func SetActiveW(platformID PlatformID, repID RepID, active bool) error {
    key := mycrypto.Hash256([]byte("active_state"), []byte(platformID), []byte(repID))
    value := []byte(strconv.FormatBool(active))
    return wasm_db.Put(key, value, nil)
}

func GetGroupW(platformID PlatformID, groupID GroupID) []RepID {
    ret := make([]RepID, 0)
    key := mycrypto.Hash256([]byte("group"), []byte(platformID), []byte(strconv.FormatInt(int64(groupID), 10)))
    value, err := wasm_db.Get(key, nil)
    if err != nil {
        return ret
    }
    json.Unmarshal(value, &ret)
    return ret
}

func SetGroupW(platformID PlatformID, groupID GroupID, repIDs []RepID) error {
    key := mycrypto.Hash256([]byte("group"), []byte(platformID), []byte(strconv.FormatInt(int64(groupID), 10)))
    value, err := json.Marshal(repIDs)
    if err != nil {
        return err
    }
    return wasm_db.Put(key, value, nil)
}

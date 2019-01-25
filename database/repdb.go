package database

import (
    "strconv"
    "BlockChainTest/mycrypto"
)

func GetTracer(platformID, repID string) []byte {
    key := mycrypto.Hash256([]byte("tracer"), []byte(platformID), []byte(repID))
    return db.get(key)
}

func PutTracer(platformID, repID string, value []byte) error {
    chainId, _ := strconv.ParseInt(platformID, 10, 64)
    key := mycrypto.Hash256([]byte("tracer"), []byte(platformID), []byte(repID))
    return db.put(key, value, chainId)
}

func IsActive(platformID, repID string) []byte {
    key := mycrypto.Hash256([]byte("active_state"), []byte(platformID), []byte(repID))
    return db.get(key)
}

func SetActive(platformID, repID string, value []byte) error {
    chainId, _ := strconv.ParseInt(platformID, 10, 64)
    key := mycrypto.Hash256([]byte("active_state"), []byte(platformID), []byte(repID))
    return db.put(key, value, chainId)
}

func GetGroup(platformID string, groupID uint64) []byte {
    key := mycrypto.Hash256([]byte("group"), []byte(platformID), []byte(strconv.FormatInt(int64(groupID), 10)))
    value := db.get(key)
    if value == nil {
        value = []byte("[]")
    }
    return value
}

func PutGroup(platformID string, groupID uint64, value []byte) error {
    chainId, _ := strconv.ParseInt(platformID, 10, 64)
    key := mycrypto.Hash256([]byte("group"), []byte(platformID), []byte(strconv.FormatInt(int64(groupID), 10)))
    return db.put(key, value, chainId)
}
package service

import "sync"

var (
    forwardedTrans = make(map[string]bool)
    transLock sync.Mutex
    forwardedBlocks = make(map[string]bool)
    blockLock sync.Mutex
)

func ForwardTransaction(id string) {
    transLock.Lock()
    defer transLock.Unlock()
    forwardedTrans[id] = true
}

func ForwardedTransaction(id string) bool {
    transLock.Lock()
    defer transLock.Unlock()
    v, exists := forwardedTrans[id]
    return v && exists
}

func ForwardBlock(id string) {
    blockLock.Lock()
    defer blockLock.Unlock()
    forwardedBlocks[id] = true
}

func ForwardedBlock(id string) bool {
    // TODO first check db and second check the pool
    blockLock.Lock()
    defer blockLock.Unlock()
    v, exists := forwardedBlocks[id]
    return v && exists
}
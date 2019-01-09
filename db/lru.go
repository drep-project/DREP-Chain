package db

import (
    "container/list"
    "errors"
)

type EvictCallback func(key string, value interface{})

type LRU struct {
    size      int
    evictList *list.List
    items     map[string]*list.Element
    onEvict   EvictCallback
}

type entry struct {
    key   string
    value interface{}
}

func NewLRU(size int, onEvict EvictCallback) (*LRU, error) {
    if size <= 0 {
        return nil, errors.New("must provide a positive size")
    }
    c := &LRU{
        size:      size,
        evictList: list.New(),
        items:     make(map[string]*list.Element),
        onEvict:   onEvict,
    }
    return c, nil
}

func (c *LRU) Purge() {
    for k, v := range c.items {
        if c.onEvict != nil {
            c.onEvict(k, v.Value.(*entry).value)
        }
        delete(c.items, k)
    }
    c.evictList.Init()
}

func (c *LRU) Add(key string, value interface{}) bool {
    if ent, ok := c.items[key]; ok {
        c.evictList.MoveToFront(ent)
        ent.Value.(*entry).value = value
        return false
    }
    ent := &entry{key, value}
    entry := c.evictList.PushFront(ent)
    c.items[key] = entry
    evict := c.evictList.Len() > c.size
    if evict {
        c.removeOldest()
    }
    return evict
}

func (c *LRU) Get(key string) (value interface{}, ok bool) {
    if ent, ok := c.items[key]; ok {
        c.evictList.MoveToFront(ent)
        return ent.Value.(*entry).value, true
    }
    return
}

// Check if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *LRU) Contains(key string) (ok bool) {
    _, ok = c.items[key]
    return ok
}

func (c *LRU) Peek(key string) (value interface{}, ok bool) {
    if ent, ok := c.items[key]; ok {
        return ent.Value.(*entry).value, true
    }
    return nil, ok
}

func (c *LRU) Remove(key string) bool {
    if ent, ok := c.items[key]; ok {
        c.removeElement(ent)
        return true
    }
    return false
}

func (c *LRU) RemoveOldest() (interface{}, interface{}, bool) {
    ent := c.evictList.Back()
    if ent != nil {
        c.removeElement(ent)
        kv := ent.Value.(*entry)
        return kv.key, kv.value, true
    }
    return nil, nil, false
}

func (c *LRU) GetOldest() (interface{}, interface{}, bool) {
    ent := c.evictList.Back()
    if ent != nil {
        kv := ent.Value.(*entry)
        return kv.key, kv.value, true
    }
    return nil, nil, false
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LRU) Keys() []interface{} {
    keys := make([]interface{}, len(c.items))
    i := 0
    for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
        keys[i] = ent.Value.(*entry).key
        i++
    }
    return keys
}

// Len returns the number of items in the cache.
func (c *LRU) Len() int {
    return c.evictList.Len()
}

// removeOldest removes the oldest item from the cache.
func (c *LRU) removeOldest() {
    ent := c.evictList.Back()
    if ent != nil {
        c.removeElement(ent)
    }
}

// removeElement is used to remove a given list element from the cache
func (c *LRU) removeElement(e *list.Element) {
    c.evictList.Remove(e)
    kv := e.Value.(*entry)
    delete(c.items, kv.key)
    if c.onEvict != nil {
        c.onEvict(kv.key, kv.value)
    }
}


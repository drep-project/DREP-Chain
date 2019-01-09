package db

import (
    "errors"
)

type dbInterface interface {
    Put(key []byte, value []byte) error
    Get(key []byte) ([]byte, error)
    Delete(key []byte) error
    BeginTransaction() Transactional
}

type Transactional interface {
    dbInterface
    Commit() error
    Discard() error
}

var (
    ErrFinished = errors.New("transaction already finished")
    ErrNotExist = errors.New("not exist")
)

type cacheEntry struct {
    key []byte
    val []byte
}

type Transaction struct {
    parent   dbInterface
    finished bool
    cache    map[string]*cacheEntry
}

func NewTransaction(parent dbInterface) *Transaction {
    return &Transaction{
        parent:     parent,
        finished:   false,
        cache:      make(map[string]*cacheEntry),
    }
}

func (t *Transaction) Put(key []byte, value []byte) error {
    if t.finished {
        return ErrFinished
    }
    t.cache[string(key)] = &cacheEntry{key: key, val: value}
    return nil
}

func (t *Transaction) Get(key []byte) ([]byte, error) {
    if t.finished {
        return nil, ErrFinished
    }
    if v, exists := t.cache[string(key)]; exists {
        return v.val, nil
    } else {
        return t.parent.Get(key)
    }
}

func (t *Transaction) Delete(key []byte) error {
    if t.finished {
        return ErrFinished
    }
    t.cache[string(key)] = &cacheEntry{key: key}
    return nil
}

func (t *Transaction) Commit() error {
    if t.finished {
        return ErrFinished
    }
    t.finished = true
    for _, v := range t.cache {
        if v == nil {
            if err := t.parent.Delete(v.key); err != nil {
                return err
            }
        } else {
            if err := t.parent.Put(v.key, v.val); err != nil {
                return err
            }
        }
    }
    return nil
}

func (t *Transaction) Discard() error {
    if t.finished {
        return ErrFinished
    }
    t.finished = true
    return nil
}

func (t *Transaction) BeginTransaction() Transactional {
    return NewTransaction(t)
}
package database

import (
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"sync"
)
type IStore interface {
	Get([]byte) ([]byte, error)
	Put([]byte,[]byte)  error
	Delete([]byte)  error
	RevertState(state *sync.Map)
    CopyState() *sync.Map

NewIterator(key []byte) Iterator
}

type Iterator iterator.Iterator
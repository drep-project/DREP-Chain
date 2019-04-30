package database

import (
	"github.com/syndtr/goleveldb/leveldb/iterator"
)
type IStore interface {
	Get([]byte) ([]byte, error)
	Put([]byte,[]byte)  error
	Delete([]byte)  error
	NewIterator(key []byte) Iterator
}

type Iterator iterator.Iterator
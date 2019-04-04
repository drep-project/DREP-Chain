package database

import (
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
	"strconv"
	"sync"
	"github.com/drep-project/binary"
)

type Database struct {
	db     *leveldb.DB
	temp   map[string][]byte
	states map[string]*State
	stores map[string]*chainTypes.Storage
	//trie  Trie
	root   []byte
	txLock sync.Mutex
}

type journal struct {
	Op       string
	Key      []byte
	Value    []byte
	Previous []byte
}

func NewDatabase(dbPath string) (*Database, error) {
	ldb, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, err
	}
	db := &Database{
		db:     ldb,
		temp:   nil,
		states: nil,
	}
	err = db.initState()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *Database) initState() error {
	err := db.db.Put([]byte("journal_depth"), new(big.Int).Bytes(), nil)
	if err != nil {
		return err
	}
	db.root = sha3.Hash256([]byte("state rootState"))
	value, err := db.get(db.root, false)
	if value != nil {
		return nil
	}
	rootState := &State{
		Sequence: "",
		Value:    []byte{0},
		IsLeaf:   true,
	}
	value, err = binary.Marshal(rootState)
	if err != nil {
		return err
	}
	return db.put(db.root, value, false)
}

func (db *Database) get(key []byte, transactional bool) ([]byte, error) {
	if !transactional {
		return db.db.Get(key, nil)
	}
	hk := bytes2Hex(key)
	value, ok := db.temp[hk]
	if !ok {
		var err error
		value, err = db.db.Get(key, nil)
		if err != nil {
			return nil, err
		}
		db.temp[hk] = value
	}
	return value, nil
}

func (db *Database) put(key []byte, value []byte, temporary bool) error {
	if !temporary {
		depthVal, err := db.db.Get([]byte("journal_depth"), nil)
		if err != nil {
			return err
		}
		var depth = new(big.Int).SetBytes(depthVal).Int64() + 1
		previous, _ := db.get(key, temporary)
		j := &journal{
			Op:       "put",
			Key:      key,
			Value:    value,
			Previous: previous,
		}
		err = db.db.Put(key, value, nil)
		if err != nil {
			return err
		}
		jVal, err := binary.Marshal(j)
		if err != nil {
			return err
		}
		err = db.db.Put([]byte("journal_" + strconv.FormatInt(depth, 10)), jVal, nil)
		if err != nil {
			return err
		}
		return db.db.Put([]byte("journal_depth"), new(big.Int).SetInt64(depth).Bytes(), nil)
	}
	db.temp[bytes2Hex(key)] = value
	return nil
}

func (db *Database) delete(key []byte, temporary bool) error {
	if !temporary {
		depthVal, err := db.db.Get([]byte("journal_depth"), nil)
		if err != nil {
			return err
		}
		var depth = new(big.Int).SetBytes(depthVal).Int64() + 1
		previous, _ := db.get(key, temporary)
		j := &journal{
			Op:       "del",
			Key:      key,
			Previous: previous,
		}
		err = db.db.Delete(key, nil)
		if err != nil {
			return err
		}
		jVal, err := binary.Marshal(j)
		if err != nil {
			return err
		}
		err = db.db.Put([]byte("journal_" + strconv.FormatInt(depth, 10)), jVal, nil)
		if err != nil {
			return err
		}
		err = db.db.Put([]byte("journal_depth"), new(big.Int).SetInt64(depth).Bytes(), nil)
		if err != nil {
			return err
		}
		return db.db.Delete(key, nil)
	}
	db.temp[bytes2Hex(key)] = nil
	return nil
}

func (db *Database) BeginTransaction() {
	db.txLock.Lock()
	db.temp = make(map[string][]byte)
	db.states = make(map[string]*State)
	db.stores = make(map[string]*chainTypes.Storage)
}

func (db *Database) EndTransaction() {
	db.temp = nil
	db.states = nil
	db.stores = nil
	db.txLock.Unlock()
}


func (db *Database) Commit() error {
	for key, value := range db.temp {
		bk := hex2Bytes(key)
		if value != nil {
			err := db.put(bk, value, false)
			if err != nil {
				return err
			}
		} else {
			err := db.delete(bk, false)
			if err != nil {
				return err
			}
		}
	}
	db.EndTransaction()
	return nil
}

func (db *Database) Discard() {
	db.EndTransaction()
}

func (db *Database) Rollback(index int64) error {
	depthVal, err := db.db.Get([]byte("journal_depth"), nil)
	if err != nil {
		return err
	}
	var depth = new(big.Int).SetBytes(depthVal).Int64()

	for i := depth; i > index; i-- {
		key := []byte("journal_" + strconv.FormatInt(i, 10))
		jVal, err := db.db.Get(key, nil)
		if err != nil {
			return err
		}
		j := &journal{}
		err = binary.Unmarshal(jVal, j)
		if err != nil {
			return err
		}
		if j.Op == "put" {
			if j.Previous == nil {
				err = db.db.Delete(j.Key, nil)
				if err != nil {
					return err
				}
			} else {
				err = db.db.Put(j.Key, j.Previous, nil)
				if err != nil {
					return err
				}
			}
		}
		if j.Op == "del" {
			err = db.db.Put(j.Key, j.Previous, nil)
			if err != nil {
				return err
			}
		}
		err = db.db.Delete(key, nil)
		if err != nil {
			return err
		}
		err = db.db.Put([]byte("journal_depth"), new(big.Int).SetInt64(index).Bytes(), nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *Database) Rollback2Block(height int64) error {
	indexVal, err := db.db.Get([]byte("depth_of_height_" + strconv.FormatInt(height, 10)), nil)
	if err != nil {
		return err
	}
	index := new(big.Int).SetBytes(indexVal).Int64()
	return db.Rollback(index)
}

func (db *Database) RecordBlockJournal(height int64) error {
	depthVal, err := db.db.Get([]byte("journal_depth"), nil)
	if err != nil {
		return err
	}
	depth := new(big.Int).SetBytes(depthVal).Int64()
	return db.db.Put([]byte("depth_of_height_" + strconv.FormatInt(height, 10)), new(big.Int).SetInt64(depth).Bytes(), nil)
}

func (db *Database) getStateRoot() []byte {
	state, _ := db.getState(db.root)
	return state.Value
}

func (db *Database) getState(key []byte) (*State, error) {
	var state *State
	hk := bytes2Hex(key)
	state, ok := db.states[hk]
	if ok {
		return state, nil
	}
	b, err := db.get(key, true)
	if err != nil {
		return nil, err
	}
	state = &State{}
	err = binary.Unmarshal(b, state)
	if err != nil {
		return nil, err
	}
	state.db = db
	db.states[hk] = state
	return state, nil
}

func (db *Database) putState(key []byte, state *State) error {
	b, err := binary.Marshal(state)
	if err != nil {
		return err
	}
	err = db.put(key, b, true)
	if err != nil {
		return err
	}
	state.db = db
	db.states[bytes2Hex(key)] = state
	return err
}

func (db *Database) delState(key []byte) error {
	err := db.delete(key, true)
	if err != nil {
		return err
	}
	db.states[bytes2Hex(key)] = nil
	return nil
}


func (db *Database) existStorage(accountName string) bool{
	key := sha3.Hash256([]byte("storage_" + accountName))
	_, err := db.get(key, false)
	if err != nil {
		return false
	}
	return true

}

func (db *Database) getStorage(accountName string) (*chainTypes.Storage, error) {
	key := sha3.Hash256([]byte("storage_" + accountName))
	value, err := db.get(key, false)
	if err != nil {
		return nil, err
	}
	storage := &chainTypes.Storage{}
	err = binary.Unmarshal(value, storage)
	if err != nil {
		return nil, err
	}
	return storage, nil
}

func (db *Database) putStorage(accountName string, storage *chainTypes.Storage) error {
	key := sha3.Hash256([]byte("storage_" + accountName))
	value, err := binary.Marshal(storage)
	if err != nil {
		return err
	}
	return db.put(key, value, true)
}

package database

import (
	"fmt"
	"github.com/drep-project/binary"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/vishalkuo/bimap"
	"math/big"
	"strconv"
	"sync"
)

type Database struct {
	db           *leveldb.DB
	temp         map[string][]byte
	states       map[string]*State
	stores       map[string]*chainTypes.Storage
	aliasAddress *bimap.BiMap //地址--别名map
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

const (
	dbOperaterMaxSeqKey = "operateMaxSeq"       //记录数据库操作的最大序列号
	maxSeqOfBlockKey    = "seqOfBlockHeight"    //块高度对应的数据库操作最大序列号
	dbOperaterJournal   = "addrOperatesJournal" //每一次数据读写过程的记录
	addressStorage      = "addressStorage"      //以地址作为KEY的对象存储
	stateRoot           = "state rootState"
)

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
	db.root = sha3.Hash256([]byte(stateRoot))
	value, _ := db.get(db.root, false)
	if value != nil {
		return nil
	}

	err := db.db.Put([]byte(dbOperaterMaxSeqKey), new(big.Int).Bytes(), nil)
	if err != nil {
		return err
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
		seqVal, err := db.db.Get([]byte(dbOperaterMaxSeqKey), nil)
		if err != nil {
			return err
		}

		var seq = new(big.Int).SetBytes(seqVal).Int64() + 1
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
		//存储seq-operater kv对
		err = db.db.Put([]byte(dbOperaterJournal+strconv.FormatInt(seq, 10)), jVal, nil)
		if err != nil {
			return err
		}
		//记录当前最高的seq
		return db.db.Put([]byte(dbOperaterMaxSeqKey), new(big.Int).SetInt64(seq).Bytes(), nil)
	}
	db.temp[bytes2Hex(key)] = value
	return nil
}

func (db *Database) delete(key []byte, temporary bool) error {
	if !temporary {
		seqVal, err := db.db.Get([]byte(dbOperaterMaxSeqKey), nil)
		if err != nil {
			return err
		}
		var seq = new(big.Int).SetBytes(seqVal).Int64() + 1
		fmt.Println("del operate seq：", seq)
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
		err = db.db.Put([]byte(dbOperaterJournal+strconv.FormatInt(seq, 10)), jVal, nil)
		if err != nil {
			return err
		}
		err = db.db.Put([]byte(dbOperaterMaxSeqKey), new(big.Int).SetInt64(seq).Bytes(), nil)
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
	db.aliasAddress = bimap.NewBiMap() //make(map[string][]byte)
}

func (db *Database) EndTransaction() {
	db.temp = nil
	db.states = nil
	db.stores = nil
	db.aliasAddress = nil
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

	err := db.aliasCommit()
	if err != nil {
		return err
	}

	db.EndTransaction()
	return nil
}

func (db *Database) Discard() {
	db.EndTransaction()
}

func (db *Database) rollback(maxBlockSeq int64, maxSeqKey, journalKey string) (error, int64) {
	seqVal, err := db.db.Get([]byte(maxSeqKey), nil)
	if err != nil {
		return err, 0
	}
	var seq = new(big.Int).SetBytes(seqVal).Int64()
	for i := seq; i > maxBlockSeq; i-- {
		key := []byte(journalKey + strconv.FormatInt(i, 10))
		jVal, err := db.db.Get(key, nil)
		if err != nil {
			return err, 0
		}
		j := &journal{}
		err = binary.Unmarshal(jVal, j)
		if err != nil {
			return err, 0
		}
		if j.Op == "put" {
			if j.Previous == nil {
				err = db.db.Delete(j.Key, nil)
				if err != nil {
					return err, 0
				}
			} else {
				err = db.db.Put(j.Key, j.Previous, nil)
				if err != nil {
					return err, 0
				}
			}
		}
		if j.Op == "del" {
			err = db.db.Put(j.Key, j.Previous, nil)
			if err != nil {
				return err, 0
			}
		}
		err = db.db.Delete(key, nil)
		if err != nil {
			return err, 0
		}
		err = db.db.Put([]byte(maxSeqKey), new(big.Int).SetInt64(maxBlockSeq).Bytes(), nil)
		if err != nil {
			return err, 0
		}
	}
	return nil, seq - maxBlockSeq
}

func (db *Database) Rollback2Block(height uint64) (error, int64) {
	return db.rollback2Block(height, maxSeqOfBlockKey)
}

func (db *Database) rollback2Block(height uint64, maxSeqOfBlock string) (error, int64) {
	value, err := db.db.Get([]byte(maxSeqOfBlock+strconv.FormatUint(height, 10)), nil)
	if err != nil {
		return err, 0
	}
	maxbockSeq := new(big.Int).SetBytes(value).Int64()

	return db.rollback(maxbockSeq, dbOperaterMaxSeqKey, dbOperaterJournal)
}

//存储height-seq kv对
func (db *Database) recordBlockJournal(height uint64, maxSeq, blockHeigthKey string) error {
	seqVal, err := db.db.Get([]byte(maxSeq), nil)
	if err != nil {
		return err
	}
	seq := new(big.Int).SetBytes(seqVal).Int64()
	return db.db.Put([]byte(blockHeigthKey+strconv.FormatUint(height, 10)), new(big.Int).SetInt64(seq).Bytes(), nil)
}

func (db *Database) RecordBlockJournal(height uint64) error {
	err := db.recordBlockJournal(height, dbOperaterMaxSeqKey, maxSeqOfBlockKey)
	if err != nil {
		return err
	}
	return nil
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

func (db *Database) getStorage(addr *crypto.CommonAddress) *chainTypes.Storage {
	storage := &chainTypes.Storage{}
	key := sha3.Hash256([]byte(addressStorage + addr.Hex()))
	value, err := db.get(key, false)
	if err != nil {
		return storage
	}
	binary.Unmarshal(value, storage)
	return storage
}

func (db *Database) putStorage(addr *crypto.CommonAddress, storage *chainTypes.Storage) error {
	key := sha3.Hash256([]byte(addressStorage + addr.Hex()))
	value, err := binary.Marshal(storage)
	if err != nil {
		return err
	}
	return db.put(key, value, true)
}

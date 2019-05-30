package database

import (
	oriBinary "encoding/binary"
	"encoding/hex"
	"github.com/drep-project/binary"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"math/big"
	"strconv"
	"sync"
)

type Database struct {
	store  IStore
	states *sync.Map
	//trie  Trie
	root          []byte
	isTransaction bool
}

var (
	aliasPrefix         = "alias"
	dbOperaterMaxSeqKey = "operateMaxSeq"       //记录数据库操作的最大序列号
	maxSeqOfBlockKey    = []byte("seqOfBlockHeight")    //块高度对应的数据库操作最大序列号
	dbOperaterJournal   = "addrOperatesJournal" //每一次数据读写过程的记录
	addressStorage      = "addressStorage"      //以地址作为KEY的对象存储
	stateRoot           = "state rootState"
)

func NewDatabase(dbPath string) (*Database, error) {
	ldb, err := NewLdbStore(dbPath)
	if err != nil {
		return nil, err
	}

	db := &Database{
		store:  ldb,
		states: new(sync.Map),
	}
	err = db.initState()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func DatabaseFromStore(store IStore) (*Database, error) {
	db := &Database{
		store:  store,
		states: new(sync.Map),
	}
	err := db.initState()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *Database) initState() error {
	db.root = sha3.Keccak256([]byte(stateRoot))
	value, _ := db.store.Get(db.root)
	if value != nil {
		return nil
	}

	err := db.store.Put([]byte(dbOperaterMaxSeqKey), new(big.Int).Bytes())
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
	return db.store.Put(db.root, value)
}

func (db *Database) rollback(maxBlockSeq int64, maxSeqKey, journalKey string) (error, int64) {
	seqVal, err := db.store.Get([]byte(maxSeqKey))
	if err != nil {
		return err, 0
	}
	var seq = new(big.Int).SetBytes(seqVal).Int64()
	for i := seq; i > maxBlockSeq; i-- {
		key := []byte(journalKey + strconv.FormatInt(i, 10))
		jVal, err := db.store.Get(key)
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
				err = db.store.Delete(j.Key)
				if err != nil {
					return err, 0
				}
			} else {
				err = db.store.Put(j.Key, j.Previous)
				if err != nil {
					return err, 0
				}
			}
		}
		if j.Op == "del" {
			err = db.store.Put(j.Key, j.Previous)
			if err != nil {
				return err, 0
			}
		}
		err = db.store.Delete(key)
		if err != nil {
			return err, 0
		}
		err = db.store.Put([]byte(maxSeqKey), new(big.Int).SetInt64(maxBlockSeq).Bytes())
		if err != nil {
			return err, 0
		}
	}
	return nil, seq - maxBlockSeq
}

func (db *Database) PutChainState(chainState *chainTypes.BestState) error {
	key := ChainStatePrefix
	value, err := binary.Marshal(chainState)
	if err != nil {
		return err
	}
	return db.Put(key, value)
}

func (db *Database) GetChainState() *chainTypes.BestState {
	key := ChainStatePrefix
	value, err := db.Get(key)
	if err != nil {
		return nil
	}
	state := &chainTypes.BestState{}
	binary.Unmarshal(value, state)
	return state
}

func (db *Database) Rollback2Block(height uint64) (error, int64) {
	keyLen := len(maxSeqOfBlockKey) + 64
	key := make([]byte, keyLen)
	copy(key, maxSeqOfBlockKey)
	oriBinary.BigEndian.PutUint64(key[len(maxSeqOfBlockKey):], height)
	value, err := db.store.Get(key)
	if err != nil {
		return err, 0
	}
	maxbockSeq := new(big.Int).SetBytes(value).Int64()

	return db.rollback(maxbockSeq, dbOperaterMaxSeqKey, dbOperaterJournal)
}

func (db *Database) RecordBlockJournal(height uint64) error {
	seqVal, err := db.store.Get([]byte(dbOperaterMaxSeqKey))
	if err != nil {
		return err
	}
	seq := new(big.Int).SetBytes(seqVal).Int64()
	keyLen := len(maxSeqOfBlockKey) + 64
	key := make([]byte, keyLen)
	copy(key, maxSeqOfBlockKey)
	oriBinary.BigEndian.PutUint64(key[len(maxSeqOfBlockKey):], height)
	return db.store.Put(key, new(big.Int).SetInt64(seq).Bytes())
}

func (db *Database) GetBlockJournal() uint64 {
	ref := db.store.NewIterator([]byte(maxSeqOfBlockKey))
	ref.Last()
	heightBytes := ref.Key()[len(maxSeqOfBlockKey):]
	height := oriBinary.BigEndian.Uint64(heightBytes)
	return height
}

func (db *Database) GetStateRoot() []byte {
	state, _ := db.GetState(db.root)
	return state.Value
}

func (db *Database) GetState(key []byte) (*State, error) {
	val, ok := db.states.Load(string(key))
	if ok && val != nil {
		return val.(*State), nil
	}
	b, err := db.store.Get(key)
	if err != nil {
		return nil, err
	}
	state := &State{}
	err = binary.Unmarshal(b, state)
	if err != nil {
		return nil, err
	}
	state.db = db
	db.states.Store(string(key), state)
	return state, nil
}

func (db *Database) PutState(key []byte, state *State) error {
	b, err := binary.Marshal(state)
	if err != nil {
		return err
	}
	err = db.store.Put(key, b)
	if err != nil {
		return err
	}
	state.db = db
	db.states.Store(string(key), state)
	return err
}

func (db *Database) DelState(key []byte) error {
	err := db.store.Delete(key)
	if err != nil {
		return err
	}
	db.states.Store(string(key), nil)
	return nil
}

func (db *Database) GetStorage(addr *crypto.CommonAddress) *chainTypes.Storage {
	storage := &chainTypes.Storage{}
	key := sha3.Keccak256([]byte(addressStorage + addr.Hex()))
	value, err := db.store.Get(key)
	if err != nil {
		return storage
	}
	binary.Unmarshal(value, storage)
	return storage
}

func (db *Database) AliasPut(key, value []byte) error {
	return db.store.Put(key, value)
}

func (db *Database) Get(key []byte) ([]byte, error) {
	return db.store.Get(key)
}

func (db *Database) Put(key []byte, value []byte) error {
	return db.store.Put(key, value)
}

func (db *Database) Delete(key []byte) error {
	return db.store.Delete(key)
}

func (db *Database) PutStorage(addr *crypto.CommonAddress, storage *chainTypes.Storage) error {
	key := sha3.Keccak256([]byte(addressStorage + addr.Hex()))
	value, err := binary.Marshal(storage)
	if err != nil {
		return err
	}
	err = db.Put(key, value)
	if err != nil {
		return err
	}

	seq := bytes2Hex(key)
	val := sha3.Keccak256(value)
	insert(db, seq, db.root, val)
	return nil
}

func (db *Database) GetBalance(addr *crypto.CommonAddress) *big.Int {
	storage := db.GetStorage(addr)

	if storage == nil {
		return new(big.Int)
	}
	return &storage.Balance
}

func (db *Database) PutBalance(addr *crypto.CommonAddress, balance *big.Int) error {
	storage := db.GetStorage(addr)
	if storage == nil {
		return ErrNoStorage
	}
	storage.Balance = *balance
	return db.PutStorage(addr, storage)
}

func (db *Database) GetNonce(addr *crypto.CommonAddress) uint64 {
	storage := db.GetStorage(addr)
	if storage == nil {
		return 0
	}
	return storage.Nonce
}

func (db *Database) PutNonce(addr *crypto.CommonAddress, nonce uint64) error {
	storage := db.GetStorage(addr)
	if storage == nil {
		return ErrNoStorage
	}
	storage.Nonce = nonce
	return db.PutStorage(addr, storage)
}

func (db *Database) GetStorageAlias(addr *crypto.CommonAddress) string {
	storage := db.GetStorage(addr)
	if storage == nil {
		return ""
	}
	return storage.Alias
}

func (db *Database) setStorageAlias(addr *crypto.CommonAddress, alias string) error {
	storage := db.GetStorage(addr)
	if storage == nil {
		return ErrNoStorage
	}
	storage.Alias = alias
	return db.PutStorage(addr, storage)
}

func (db *Database) AliasSet(addr *crypto.CommonAddress, alias string) (err error) {
	if alias != "" {
		//1 检查别名是否存在
		b := db.AliasExist(alias)
		if b {
			return ErrUsedAlias
		}

		//2 存入以alias为key的k-v对
		err = db.AliasPut([]byte(aliasPrefix+alias), addr.Bytes())
	} else {
		return ErrInvalidateAlias
	}

	if err != nil {
		return err
	}

	//put to stroage
	err = db.setStorageAlias(addr, alias)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) AliasGet(alias string) *crypto.CommonAddress {
	buf, err := db.store.Get([]byte(aliasPrefix + alias))
	if err != nil {
		return nil
	}
	addr := crypto.CommonAddress{}
	addr.SetBytes(buf)
	return &addr
}

func (db *Database) AliasExist(alias string) bool {
	_, err := db.store.Get([]byte(aliasPrefix + alias))
	if err != nil {
		return false
	}
	return true
}

func (db *Database) GetByteCode(addr *crypto.CommonAddress) []byte {
	storage := db.GetStorage(addr)
	if storage == nil {
		return nil
	}
	return storage.ByteCode
}

func (db *Database) PutByteCode(addr *crypto.CommonAddress, byteCode []byte) error {
	storage := db.GetStorage(addr)
	if storage == nil {
		return ErrNoStorage
	}
	storage.ByteCode = byteCode
	storage.CodeHash = crypto.GetByteCodeHash(byteCode)
	return db.PutStorage(addr, storage)
}

func (db *Database) GetCodeHash(addr *crypto.CommonAddress) crypto.Hash {
	storage := db.GetStorage(addr)
	if storage == nil {
		return crypto.Hash{}
	}
	return storage.CodeHash
}

func (db *Database) GetReputation(addr *crypto.CommonAddress) *big.Int {
	storage := db.GetStorage(addr)
	if storage == nil {
		return big.NewInt(0)
	}
	return storage.Reputation
}

func (db *Database) GetLogs(txHash []byte) []*chainTypes.Log {
	key := sha3.Keccak256([]byte("logs_" + hex.EncodeToString(txHash)))
	value, err := db.Get(key)
	if err != nil {
		return make([]*chainTypes.Log, 0)
	}
	var logs []*chainTypes.Log
	err = binary.Unmarshal(value, &logs)
	if err != nil {
		return make([]*chainTypes.Log, 0)
	}
	return logs
}

func (db *Database) PutLogs(logs []*chainTypes.Log, txHash []byte) error {
	key := sha3.Keccak256([]byte("logs_" + hex.EncodeToString(txHash)))
	value, err := binary.Marshal(logs)
	if err != nil {
		return err
	}
	return db.Put(key, value)
}

func (db *Database) AddLog(log *chainTypes.Log) error {
	logs := db.GetLogs(log.TxHash)
	logs = append(logs, log)
	return db.PutLogs(logs, log.TxHash)
}

func (db *Database) Load(x *big.Int) []byte {
	value, _ := db.Get(x.Bytes())
	return value
}

func (db *Database) Store(x, y *big.Int) error {
	return db.Put(x.Bytes(), y.Bytes())
}

func (db *Database) AddBalance(addr *crypto.CommonAddress, amount *big.Int) error {
	balance := db.GetBalance(addr)
	if balance == nil {
		balance = new(big.Int).SetInt64(0)
	}
	return db.PutBalance(addr, new(big.Int).Add(balance, amount))
}

func (db *Database) SubBalance(addr *crypto.CommonAddress, amount *big.Int) error {
	balance := db.GetBalance(addr)
	return db.PutBalance(addr, new(big.Int).Sub(balance, amount))
}

func (db *Database) PutBlock(block *chainTypes.Block) error {
	hash := block.Header.Hash()
	key := append(BlockPrefix, hash[:]...)
	value, err := binary.Marshal(block)
	if err != nil {
		return err
	}
	return db.store.Put(key, value)
}

func (db *Database) GetBlock(hash *crypto.Hash) (*chainTypes.Block, error) {
	key := append(BlockPrefix, hash[:]...)
	val ,err := db.store.Get(key)
	if err != nil {
		return nil, err
	}
	block := &chainTypes.Block{}
	err = binary.Unmarshal(val, block)
	if err != nil {
		return nil, err
	}
	return block, nil
}


func (db *Database) GetBlockNode(hash *crypto.Hash,blockHeight uint64) (*chainTypes.BlockHeader, chainTypes.BlockStatus,error) {
	key := db.blockIndexKey(hash, blockHeight)
	value, err := db.Get(key)
	if err != nil {
		return nil, 0, err
	}
	blockHeader := &chainTypes.BlockHeader{}
	binary.Unmarshal(value[0:len(value)-1], blockHeader)
	status := value[len(value)-1:len(value)][0]
	return blockHeader, chainTypes.BlockStatus(status), nil
}


func (db *Database)  PutBlockNode(blockNode *chainTypes.BlockNode) error {
	header := blockNode.Header()
	value, err := binary.Marshal(header)
	if err != nil {
		return err
	}
	key := db.blockIndexKey(blockNode.Hash, blockNode.Height)
	value = append(value, byte(blockNode.Status))
	return db.store.Put(key, value)
}

func (db *Database)  blockIndexKey(blockHash *crypto.Hash, blockHeight uint64) []byte {
	indexKey := make([]byte, len(BlockNodePrefix)+crypto.HashLength+8)
	copy(indexKey[0:len(BlockNodePrefix)], BlockNodePrefix[:])
	binary.BigEndian.PutUint64(indexKey[len(BlockNodePrefix):len(BlockNodePrefix)+8], uint64(blockHeight))
	copy(indexKey[len(BlockNodePrefix)+8:len(BlockNodePrefix)+40], blockHash[:])
	return indexKey
}

func (db *Database) BeginTransaction() *Database {
	return &Database{
		store:         NewTransactionStore(db.store),
		states:        new (sync.Map),
		root:          db.root,
		isTransaction: true,
	}
}

func (db *Database) Commit(needLog bool) {
	if db.isTransaction {
		db.store.(*TransactionStore).Flush(needLog)
	}
}

func (db *Database) Discard() {
	if db.isTransaction {
		db.store.(*TransactionStore).Clear()
	}
}

func (db *Database) RevertState(shot *SnapShot) {
	db.states = shot.StateShot
	db.store.RevertState(shot.StoreShot)
}

func (db *Database) CopyState() *SnapShot {
	newStoreShot :=db.store.CopyState()
	newStateShot := copyMap(db.states)
	return &SnapShot{newStoreShot, newStateShot}
}

func copyMap(m *sync.Map) *sync.Map {
	newMap := new (sync.Map)
	m.Range(func(key, value interface{}) bool {
		newMap.Store(key, value)
		return true
	})
	return newMap
}


type SnapShot struct {
	StoreShot *sync.Map
	StateShot *sync.Map
}


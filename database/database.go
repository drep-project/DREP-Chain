package database

import (
	"github.com/drep-project/binary"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/database/drepdb"
	"github.com/drep-project/drep-chain/database/drepdb/leveldb"
	"github.com/drep-project/drep-chain/database/trie"

	oriBinary "encoding/binary"
	"fmt"
	"math/big"
	"strconv"
	"sync"
)

type Database struct {
	diskDb drepdb.KeyValueStore //实际磁盘数据库
	cache  *TransactionStore    //缓存数据库，调用flush才会把数据写入到diskDb中
	trie   *trie.SecureTrie     //全局状态树
	trieDb *trie.Database       //状态树存储到磁盘时，使用到的db
}

var (
	aliasPrefix         = "alias"
	dbOperaterMaxSeqKey = "operateMaxSeq"            //记录数据库操作的最大序列号
	maxSeqOfBlockKey    = []byte("seqOfBlockHeight") //块高度对应的数据库操作最大序列号
	dbOperaterJournal   = "addrOperatesJournal"      //每一次数据读写过程的记录
	addressStorage      = "addressStorage"           //以地址作为KEY的对象存储
)

func NewDatabase(dbPath string) (*Database, error) {
	diskDb, err := leveldb.New(dbPath, 16, 512, "")
	db := &Database{
		diskDb: diskDb,
	}

	err = db.initState()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func DatabaseFromStore(diskDb drepdb.KeyValueStore) (*Database, error) {
	db := &Database{
		diskDb: diskDb,
	}

	err := db.initState()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *Database) initState() error {
	db.trieDb = trie.NewDatabaseWithCache(db.diskDb, 0)

	var err error
	value, _ := db.diskDb.Get(trie.EmptyRoot[:])
	if value == nil {
		db.diskDb.Put([]byte(dbOperaterMaxSeqKey), new(big.Int).Bytes())
		db.trie, err = trie.NewSecure(crypto.Hash{}, db.trieDb)
		db.diskDb.Put(trie.EmptyRoot[:], []byte{0})
	} else {
		chainState := db.GetChainState()
		journalHeight := db.GetBlockJournal()
		header, _, err := db.GetBlockNode(&chainState.Hash, journalHeight)
		if err != nil {
			return err
		}
		db.trie, err = trie.NewSecure(crypto.Bytes2Hash(header.StateRoot), db.trieDb)
	}

	return err
}

func (db *Database) rollback(maxBlockSeq int64, maxSeqKey, journalKey string) (error, int64) {
	seqVal, err := db.diskDb.Get([]byte(maxSeqKey))
	if err != nil {
		return err, 0
	}
	var seq = new(big.Int).SetBytes(seqVal).Int64()
	for i := seq; i > maxBlockSeq; i-- {
		key := []byte(journalKey + strconv.FormatInt(i, 10))
		jVal, err := db.diskDb.Get(key)
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
				err = db.diskDb.Delete(j.Key)
				if err != nil {
					return err, 0
				}
			} else {
				err = db.diskDb.Put(j.Key, j.Previous)
				if err != nil {
					return err, 0
				}
			}
		}
		if j.Op == "del" {
			err = db.diskDb.Put(j.Key, j.Previous)
			if err != nil {
				return err, 0
			}
		}
		err = db.diskDb.Delete(key)
		if err != nil {
			return err, 0
		}
		err = db.diskDb.Put([]byte(maxSeqKey), new(big.Int).SetInt64(maxBlockSeq).Bytes())
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
	value, err := db.diskDb.Get(key)
	if err != nil {
		return err, 0
	}
	maxbockSeq := new(big.Int).SetBytes(value).Int64()

	return db.rollback(maxbockSeq, dbOperaterMaxSeqKey, dbOperaterJournal)
}

func (db *Database) SetBlockJournal(height uint64) error {
	seqVal, err := db.Get([]byte(dbOperaterMaxSeqKey))
	if err != nil {
		return err
	}
	seq := new(big.Int).SetBytes(seqVal).Int64()
	keyLen := len(maxSeqOfBlockKey) + 64
	key := make([]byte, keyLen)
	copy(key, maxSeqOfBlockKey)
	oriBinary.BigEndian.PutUint64(key[len(maxSeqOfBlockKey):], height)
	return db.Put(key, new(big.Int).SetInt64(seq).Bytes())
}

func (db *Database) GetBlockJournal() uint64 {
	it := db.diskDb.NewIteratorWithPrefix([]byte(maxSeqOfBlockKey))

	var heightBytes []byte
	for it.Next() {
		heightBytes = it.Key()[len(maxSeqOfBlockKey):]
	}

	height := oriBinary.BigEndian.Uint64(heightBytes)
	return height
}

func (db *Database) GetStorage(addr *crypto.CommonAddress) *chainTypes.Storage {
	storage := &chainTypes.Storage{}
	key := sha3.Keccak256([]byte(addressStorage + addr.Hex()))

	var value []byte
	var err error
	if db.cache != nil {
		value, err = db.cache.Get(key)
	} else {
		value, err = db.trie.TryGet(key)
	}
	if err != nil {
		log.Errorf("get storage err:%v", err)
		return nil
	}
	binary.Unmarshal(value, storage)
	return storage
}

func (db *Database) PutStorage(addr *crypto.CommonAddress, storage *chainTypes.Storage) error {
	key := sha3.Keccak256([]byte(addressStorage + addr.Hex()))
	value, err := binary.Marshal(storage)
	if err != nil {
		return err
	}

	if db.cache != nil {
		return db.cache.Put(key, value)
	} else {
		return db.trie.TryUpdate(key, value)
	}
}

func (db *Database) AliasPut(key, value []byte) error {
	return db.diskDb.Put(key, value)
}

func (db *Database) Get(key []byte) ([]byte, error) {
	return db.diskDb.Get(key)
}

func (db *Database) Put(key []byte, value []byte) error {
	return db.diskDb.Put(key, value)
}

func (db *Database) Delete(key []byte) error {
	return db.diskDb.Delete(key)
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

//alias为key的k-v
func (db *Database) AliasGet(alias string) *crypto.CommonAddress {
	buf, err := db.diskDb.Get([]byte(aliasPrefix + alias))
	if err != nil {
		return nil
	}
	addr := crypto.CommonAddress{}
	addr.SetBytes(buf)
	return &addr
}

func (db *Database) AliasExist(alias string) bool {
	_, err := db.diskDb.Get([]byte(aliasPrefix + alias))
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

func (db *Database) GetLogs(txHash crypto.Hash) []*chainTypes.Log {
	key := sha3.Keccak256([]byte("logs_" + txHash.String()))
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

func (db *Database) PutLogs(logs []*chainTypes.Log, txHash crypto.Hash) error {
	key := sha3.Keccak256([]byte("logs_" + txHash.String()))
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

func (db *Database) PutReceipt(txHash crypto.Hash, receipt *chainTypes.Receipt) error {
	key := sha3.Keccak256([]byte("receipt_" + txHash.String()))
	fmt.Println("tx receipt: ", receipt)
	fmt.Println("detail:")
	fmt.Println("PostState: ", receipt.PostState)
	fmt.Println("Status: ", receipt.Status)
	fmt.Println("CumulativeGasUsed: ", receipt.CumulativeGasUsed)
	fmt.Println("Logs: ", receipt.Logs, receipt.Logs == nil)
	fmt.Println("TxHash: ", receipt.TxHash)
	fmt.Println("ContratAddress: ", receipt.ContractAddress)
	fmt.Println("GasUsed: ", receipt.GasUsed)
	fmt.Println("GasFee: ", receipt.GasUsed)
	fmt.Println("Ret: ", receipt.Ret, receipt.Ret == nil)
	value, err := binary.Marshal(receipt)
	fmt.Println("err11: ", err)
	if err != nil {
		return err
	}
	return db.Put(key, value)
}

func (db *Database) GetReceipt(txHash crypto.Hash) *chainTypes.Receipt {
	key := sha3.Keccak256([]byte("receipt_" + txHash.String()))
	value, err := db.Get(key)
	fmt.Println("err12: ", err)
	fmt.Println("val: ", value)
	if err != nil {
		return nil
	}
	receipt := &chainTypes.Receipt{}
	err = binary.Unmarshal(value, receipt)
	fmt.Println("err13: ", err)
	if err != nil {
		return nil
	}
	return receipt
}

func (db *Database) PutReceipts(blockHash crypto.Hash, receipts []*chainTypes.Receipt) error {
	key := sha3.Keccak256([]byte("receipts_" + blockHash.String()))
	value, err := binary.Marshal(receipts)
	if err != nil {
		return err
	}
	return db.Put(key, value)
}

func (db *Database) GetReceipts(blockHash crypto.Hash) []*chainTypes.Receipt {
	key := sha3.Keccak256([]byte("receipts_" + blockHash.String()))
	value, err := db.Get(key)
	if err != nil {
		return make([]*chainTypes.Receipt, 0)
	}
	var receipts []*chainTypes.Receipt
	err = binary.Unmarshal(value, &receipts)
	if err != nil {
		return make([]*chainTypes.Receipt, 0)
	}
	return receipts
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
	return db.Put(key, value)
}

func (db *Database) GetBlock(hash *crypto.Hash) (*chainTypes.Block, error) {
	key := append(BlockPrefix, hash[:]...)
	val, err := db.Get(key)
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

func (db *Database) GetBlockNode(hash *crypto.Hash, blockHeight uint64) (*chainTypes.BlockHeader, chainTypes.BlockStatus, error) {
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

func (db *Database) PutBlockNode(blockNode *chainTypes.BlockNode) error {
	header := blockNode.Header()
	value, err := binary.Marshal(header)
	if err != nil {
		return err
	}
	key := db.blockIndexKey(blockNode.Hash, blockNode.Height)

	value = append(value, byte(blockNode.Status))
	return db.Put(key, value)
}

func (db *Database) blockIndexKey(blockHash *crypto.Hash, blockHeight uint64) []byte {
	indexKey := make([]byte, len(BlockNodePrefix)+crypto.HashLength+8)
	copy(indexKey[0:len(BlockNodePrefix)], BlockNodePrefix[:])
	binary.BigEndian.PutUint64(indexKey[len(BlockNodePrefix):len(BlockNodePrefix)+8], uint64(blockHeight))
	copy(indexKey[len(BlockNodePrefix)+8:len(BlockNodePrefix)+40], blockHash[:])
	return indexKey
}

func (db *Database) BeginTransaction() *Database {
	return &Database{
		diskDb: db.diskDb,
		cache:  NewTransactionStore(db.trie, db.diskDb),
		trie:   db.trie,
	}
}

func (db *Database) Commit(needLog bool) {
	if db.cache != nil {
		db.cache.Flush(needLog)
	}
}

func (db *Database) Discard() {
	if db.cache != nil {
		db.cache.Clear()
	}
}

func (db *Database) RevertState(shot *SnapShot) {
	db.cache.RevertState(shot.StoreShot)
}

func (db *Database) CopyState() *SnapShot {
	newStoreShot := db.cache.CopyState()
	return &SnapShot{newStoreShot}
}

func (db *Database) GetStateRoot() []byte {
	return db.trie.Hash().Bytes()
}

func copyMap(m *sync.Map) *sync.Map {
	newMap := new(sync.Map)
	m.Range(func(key, value interface{}) bool {
		if value == nil {
			newMap.Store(key, value)
		} else {
			switch t := value.(type) {
			case []byte:
				newBytes := make([]byte, len(t))
				copy(newBytes, t)
				newMap.Store(key, newBytes)
			default:
				panic("never run here")
			}
		}
		return true
	})
	return newMap
}

type SnapShot struct {
	StoreShot *sync.Map
}

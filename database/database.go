package database

import (
	"github.com/drep-project/binary"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"github.com/drep-project/DREP-Chain/database/drepdb"
	"github.com/drep-project/DREP-Chain/database/drepdb/leveldb"
	"github.com/drep-project/DREP-Chain/database/drepdb/memorydb"
	"github.com/drep-project/DREP-Chain/database/trie"
	"github.com/drep-project/DREP-Chain/types"

	"fmt"
	"math/big"
	"strconv"
)

type Database struct {
	diskDb drepdb.KeyValueStore //实际磁盘数据库
	cache  *TransactionStore    //数据属于storage的缓存，调用flush才会把数据写入到diskDb中
	trie   *trie.SecureTrie     //全局状态树
	trieDb *trie.Database       //状态树存储到磁盘时，使用到的db
}

var (
	aliasPrefix         = "alias"
	dbOperaterMaxSeqKey = "operateMaxSeq"       //记录数据库操作的最大序列号
	dbOperaterJournal   = "addrOperatesJournal" //每一次数据读写过程的记录
	addressStorage      = "addressStorage"      //以地址作为KEY的对象存储
	candidateAddrs      = "candidateAddrs"      //参与竞选出块节点的地址集合
)

func NewDatabase(dbPath string) (*Database, error) {
	diskDb, err := leveldb.New(dbPath, 16, 512, "")
	if err != nil {
		return nil, err
	}

	db, err := DatabaseFromStore(diskDb)
	return db, nil
}

func DatabaseFromStore(diskDb drepdb.KeyValueStore) (*Database, error) {
	db := &Database{
		diskDb: diskDb,
	}

	db.trieDb = trie.NewDatabaseWithCache(db.diskDb, 0)

	err := db.initState()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *Database) RecoverTrie(root []byte) bool {
	var err error
	db.trie, err = trie.NewSecure(crypto.Bytes2Hash(root), db.trieDb)
	if err != nil {
		return false
	}
	return true
}

func (db *Database) initState() error {
	value, _ := db.diskDb.Get(trie.EmptyRoot[:])
	if value == nil {
		db.diskDb.Put([]byte(dbOperaterMaxSeqKey), new(big.Int).Bytes())

		var err error
		db.trie, err = trie.NewSecure(crypto.Hash{}, db.trieDb)
		if err != nil {
			return err
		}

		db.diskDb.Put(trie.EmptyRoot[:], []byte{0})
	}

	return nil
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

func (db *Database) Rollback2Block(height uint64, hash *crypto.Hash) (error, int64) {
	var err error

	//删除blocknode
	func() {
		key := db.blockIndexKey(hash, height)
		err = db.Delete(key)
	}()

	if err != nil {
		return err, 0
	}

	//删除block
	func() {
		key := append(BlockPrefix, hash[:]...)
		err = db.Delete(key)
	}()

	if err != nil {
		return err, 0
	}

	return nil, 0
}

func (db *Database) GetStorage(addr *crypto.CommonAddress) (*types.Storage, error) {
	storage := &types.Storage{}
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
		return nil, err
	}
	if value == nil {
		return nil, nil
	} else {
		err = binary.Unmarshal(value, storage)
		if err != nil {
			return nil, err
		}
	}
	return storage, nil
}

func (db *Database) DeleteStorage(addr *crypto.CommonAddress) error {
	key := sha3.Keccak256([]byte(addressStorage + addr.Hex()))
	if db.cache != nil {
		return db.cache.Delete(key)
	} else {
		err := db.trie.TryDelete(key)
		if err != nil {
			return err
		}
		_, err = db.trie.Commit(nil)
		return err
	}
}

func (db *Database) PutStorage(addr *crypto.CommonAddress, storage *types.Storage) error {
	key := sha3.Keccak256([]byte(addressStorage + addr.Hex()))
	value, err := binary.Marshal(storage)
	if err != nil {
		return err
	}

	if db.cache != nil {
		return db.cache.Put(key, value)
	} else {
		err = db.trie.TryUpdate(key, value)
		if err != nil {
			return err
		}
		_, err = db.trie.Commit(nil)
		return err
	}
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
	storage, _ := db.GetStorage(addr)

	if storage == nil {
		return new(big.Int)
	}
	return &storage.Balance
}

func (db *Database) PutBalance(addr *crypto.CommonAddress, balance *big.Int) error {
	storage, _ := db.GetStorage(addr)
	if storage == nil {
		storage = &types.Storage{}
	}
	storage.Balance = *balance
	return db.PutStorage(addr, storage)
}

func (db *Database) GetNonce(addr *crypto.CommonAddress) uint64 {
	storage, _ := db.GetStorage(addr)
	if storage == nil {
		return 0
	}
	return storage.Nonce
}

func (db *Database) PutNonce(addr *crypto.CommonAddress, nonce uint64) error {
	storage, _ := db.GetStorage(addr)
	if storage == nil {
		storage = &types.Storage{}
	}
	storage.Nonce = nonce
	return db.PutStorage(addr, storage)
}

func (db *Database) GetStorageAlias(addr *crypto.CommonAddress) string {
	storage, _ := db.GetStorage(addr)
	if storage == nil {
		return ""
	}
	return storage.Alias
}

func (db *Database) setStorageAlias(addr *crypto.CommonAddress, alias string) error {
	storage, _ := db.GetStorage(addr)
	if storage == nil {
		storage = &types.Storage{}
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
		err = db.AliasPut(alias, addr.Bytes())
		if err != nil {
			return err
		}
	} else {
		return ErrInvalidateAlias
	}

	//put to stroage
	err = db.setStorageAlias(addr, alias)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) AliasPut(alias string, value []byte) error {
	db.cache.Put([]byte(aliasPrefix+alias), value)
	return nil
}

//alias为key的k-v
func (db *Database) AliasGet(alias string) *crypto.CommonAddress {
	buf, err := db.trie.TryGet([]byte(aliasPrefix + alias))
	if err != nil {
		return nil
	}
	addr := crypto.CommonAddress{}
	addr.SetBytes(buf)
	return &addr
}

func (db *Database) AliasExist(alias string) bool {
	if db.cache != nil {
		_, ok := db.cache.dirties.Load(alias)
		if ok {
			return true
		}
	}
	_, err := db.diskDb.Get([]byte(aliasPrefix + alias))
	if err != nil {
		return false
	}
	return true
}

func (db *Database) GetByteCode(addr *crypto.CommonAddress) []byte {
	storage, _ := db.GetStorage(addr)
	if storage == nil {
		return nil
	}
	return storage.ByteCode
}

func (db *Database) PutByteCode(addr *crypto.CommonAddress, byteCode []byte) error {
	storage, _ := db.GetStorage(addr)
	if storage == nil {
		storage = &types.Storage{}
	}
	storage.ByteCode = byteCode
	storage.CodeHash = crypto.GetByteCodeHash(byteCode)
	return db.PutStorage(addr, storage)
}

func (db *Database) GetCodeHash(addr *crypto.CommonAddress) crypto.Hash {
	storage, _ := db.GetStorage(addr)
	if storage == nil {
		return crypto.Hash{}
	}
	return storage.CodeHash
}

func (db *Database) GetReputation(addr *crypto.CommonAddress) *big.Int {
	storage, _ := db.GetStorage(addr)
	if storage == nil {
		return big.NewInt(0)
	}
	return &storage.Reputation
}

//func (db *Database) GetLogs(txHash crypto.Hash) []*types.Log {
//	key := sha3.Keccak256([]byte("logs_" + txHash.String()))
//	value, err := db.Get(key)
//	if err != nil {
//		return make([]*types.Log, 0)
//	}
//	var logs []*types.Log
//	err = binary.Unmarshal(value, &logs)
//	if err != nil {
//		return make([]*types.Log, 0)
//	}
//	return logs
//}

func (db *Database) PutLogs(logs []*types.Log, txHash crypto.Hash) error {
	key := sha3.Keccak256([]byte("logs_" + txHash.String()))
	value, err := binary.Marshal(logs)
	if err != nil {
		return err
	}
	return db.Put(key, value)
}

//func (db *Database) AddLog(log *types.Log) error {
//	//logs := db.GetLogs(log.TxHash)
//	//logs = append(logs, log)
//	//return db.PutLogs(logs, log.TxHash)
//	return nil
//}

func (db *Database) PutReceipt(txHash crypto.Hash, receipt *types.Receipt) error {
	key := sha3.Keccak256([]byte("receipt_" + txHash.String()))
	//fmt.Println("tx receipt: ", receipt)
	//fmt.Println("detail:")
	//fmt.Println("PostState: ", receipt.PostState)
	//fmt.Println("Status: ", receipt.Status)
	//fmt.Println("CumulativeGasUsed: ", receipt.CumulativeGasUsed)
	//fmt.Println("Logs: ", receipt.Logs, receipt.Logs == nil)
	//fmt.Println("TxHash: ", receipt.TxHash)
	//fmt.Println("ContratAddress: ", receipt.ContractAddress)
	//fmt.Println("GasUsed: ", receipt.GasUsed)
	//fmt.Println("BlockNumber: ", receipt.BlockNumber)
	//fmt.Println("BlockHash: ", receipt.BlockHash)
	value, err := binary.Marshal(receipt)
	//fmt.Println("err11: ", err)
	if err != nil {
		return err
	}
	return db.Put(key, value)
}

func (db *Database) GetReceipt(txHash crypto.Hash) *types.Receipt {
	key := sha3.Keccak256([]byte("receipt_" + txHash.String()))
	value, err := db.Get(key)
	//fmt.Println("err12: ", err)
	//fmt.Println("val: ", value)
	if err != nil {
		return nil
	}
	receipt := &types.Receipt{}
	err = binary.Unmarshal(value, receipt)
	fmt.Println("err13: ", err)
	if err != nil {
		return nil
	}
	return receipt
}

func (db *Database) PutReceipts(blockHash crypto.Hash, receipts []*types.Receipt) error {
	key := sha3.Keccak256([]byte("receipts_" + blockHash.String()))
	value, err := binary.Marshal(receipts)
	if err != nil {
		return err
	}
	return db.Put(key, value)
}

func (db *Database) GetReceipts(blockHash crypto.Hash) []*types.Receipt {
	key := sha3.Keccak256([]byte("receipts_" + blockHash.String()))
	value, err := db.Get(key)
	if err != nil {
		return make([]*types.Receipt, 0)
	}
	var receipts []*types.Receipt
	err = binary.Unmarshal(value, &receipts)
	if err != nil {
		return make([]*types.Receipt, 0)
	}
	return receipts
}

func (db *Database) DeleteReceipts(blockHash crypto.Hash) error {
	key := sha3.Keccak256([]byte("receipts_" + blockHash.String()))
	return db.Delete(key)
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

func (db *Database) PutBlock(block *types.Block) error {
	hash := block.Header.Hash()
	key := append(BlockPrefix, hash[:]...)
	value, err := binary.Marshal(block)
	if err != nil {
		return err
	}
	return db.Put(key, value)
}

func (db *Database) GetBlock(hash *crypto.Hash) (*types.Block, error) {
	key := append(BlockPrefix, hash[:]...)
	val, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	block := &types.Block{}
	err = binary.Unmarshal(val, block)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (db *Database) GetBlockHeader(hash *crypto.Hash) (*types.BlockHeader, error) {
	key := append(BlockPrefix, hash[:]...)
	val, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	block := &types.Block{}
	err = binary.Unmarshal(val, block)
	if err != nil {
		return nil, err
	}
	return block.Header, nil
}

func (db *Database) FindCommonAncestor(a, b *types.BlockHeader) *types.BlockHeader {
	for bn := b.Height; a.Height > bn; {
		a, _ := db.GetBlockHeader(&a.PreviousHash)
		if a == nil {
			return nil
		}
	}
	for an := a.Height; an < b.Height; {
		b, _ := db.GetBlockHeader(&b.PreviousHash)
		if b == nil {
			return nil
		}
	}
	for a.Hash() != b.Hash() {
		a, _ := db.GetBlockHeader(&a.PreviousHash)
		if a == nil {
			return nil
		}
		b, _ := db.GetBlockHeader(&b.PreviousHash)
		if b == nil {
			return nil
		}
	}
	return a
}

func (db *Database) GetBlockNode(hash *crypto.Hash, blockHeight uint64) (*types.BlockHeader, types.BlockStatus, error) {
	key := db.blockIndexKey(hash, blockHeight)

	value, err := db.Get(key)
	if err != nil {
		return nil, 0, err
	}
	blockHeader := &types.BlockHeader{}
	binary.Unmarshal(value[0:len(value)-1], blockHeader)
	status := value[len(value)-1 : len(value)][0]
	return blockHeader, types.BlockStatus(status), nil
}

func (db *Database) PutBlockNode(blockNode *types.BlockNode) error {
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

// storeToDB 决定数据是否写入到db,如果是true，则返回的对象可以把数据写入到物理磁盘
func (db *Database) BeginTransaction(storeToDB bool) *Database {
	if !storeToDB {
		writeTrieDB := trie.NewDatabase(memorydb.New())
		newTrie, err := trie.NewSecureNewWithRWDB(db.trie.Hash(), db.trieDb, writeTrieDB)
		if err != nil {
			log.WithField("err", err).Error("NewSecure2")
			return nil
		}
		return &Database{
			diskDb: db.diskDb,
			cache:  NewTransactionStore(newTrie, db.diskDb),
			trie:   newTrie,
			trieDb: db.trieDb,
		}
	} else {
		return &Database{
			diskDb: db.diskDb,
			cache:  NewTransactionStore(db.trie, db.diskDb),
			trie:   db.trie,
			trieDb: db.trieDb,
		}
	}
}

func (db *Database) NewBatch() drepdb.Batch {
	return db.diskDb.NewBatch()
}

func (db *Database) Commit() {
	if db.cache != nil {
		db.cache.Flush()
	}
}

func (db *Database) RevertState(shot *SnapShot) {
	db.cache.RevertState(shot.storageDirties)
}

func (db *Database) CopyState() *SnapShot {
	return db.cache.CopyState()
}

func (db *Database) GetStateRoot() []byte {
	return db.trie.Hash().Bytes()
}

func (db *Database) UpdateCandidateAddr(addr *crypto.CommonAddress, add bool) error {
	//读取
	addrs, err := db.GetCandidateAddrs()
	if err != nil {
		return err
	}

	if add {
		if len(addrs) > 0 {
			addrs[*addr] = struct{}{}
		} else {
			addrs = make(map[crypto.CommonAddress]struct{})
			addrs[*addr] = struct{}{}
		}
	} else { //del
		if len(addrs) == 0 {
			return nil
		} else {
			if _, ok := addrs[*addr]; ok {
				delete(addrs, *addr)
			}
		}
	}

	addrsBuf, err := binary.Marshal(addrs)
	if err == nil {
		db.cache.Put([]byte(candidateAddrs), addrsBuf)
	}
	return err
}

func (db *Database) AddCandidateAddr(addr *crypto.CommonAddress) error {
	return db.UpdateCandidateAddr(addr, true)
}

func (db *Database) DelCandidateAddr(addr *crypto.CommonAddress) error {
	return db.UpdateCandidateAddr(addr, false)
}

func (db *Database) GetCandidateAddrs() (map[crypto.CommonAddress]struct{}, error) {
	var addrsBuf []byte
	var err error
	key := []byte(candidateAddrs)
	addrs := make(map[crypto.CommonAddress]struct{})

	if db.cache != nil {
		addrsBuf, err = db.cache.Get(key)
	} else {
		addrsBuf, err = db.trie.TryGet(key)
	}

	if err != nil {
		log.Errorf("GetCandidateAddrs:%v", err)
		return nil, err
	}

	if addrsBuf == nil {
		return nil, nil
	}

	err = binary.Unmarshal(addrsBuf, &addrs)
	if err != nil {
		log.Errorf("GetCandidateAddrs, Unmarshal:%v", err)
		return nil, err
	}
	return addrs, nil
}

type SnapShot dirtiesKV

package chain

import (
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/common/trie"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/database/dbinterface"
	"github.com/drep-project/drep-chain/types"
	"math/big"
)

const (
	aliasPrefix         = "alias"
	dbOperaterMaxSeqKey = "operateMaxSeq"       //记录数据库操作的最大序列号
	dbOperaterJournal   = "addrOperatesJournal" //每一次数据读写过程的记录
	addressStorage      = "addressStorage"      //以地址作为KEY的对象存储
)

var (
	MetaDataPrefix   = []byte("metaData_")
	ChainStatePrefix = []byte("chainState_")
	BlockPrefix      = []byte("block_")
	BlockNodePrefix  = []byte("blockNode_")
)

type TrieRead interface {
	GetStorage(addr *crypto.CommonAddress) (*types.Storage, error)
	GetStorageAlias(addr *crypto.CommonAddress) string
	AliasGet(alias string) (*crypto.CommonAddress, error)
	AliasExist(alias string) bool
	GetBalance(addr *crypto.CommonAddress) *big.Int
	GetNonce(addr *crypto.CommonAddress) uint64
	GetByteCode(addr *crypto.CommonAddress) []byte
	GetCodeHash(addr *crypto.CommonAddress) crypto.Hash
	GetReputation(addr *crypto.CommonAddress) *big.Int
}

type TrieStore struct {
	store  dbinterface.KeyValueStore
	cache  *database.TransactionStore //数据属于storage的缓存，调用flush才会把数据写入到diskDb中
	trie   *trie.SecureTrie           //全局状态树  临时树（临时变量）
	trieDb *trie.Database             //状态树存储到磁盘时，使用到的db
}

func TrieStoreFromStore(store dbinterface.KeyValueStore, stateRoot []byte) (*TrieStore, error) {
	db := &TrieStore{
		store: store,
	}

	db.trieDb = trie.NewDatabaseWithCache(db.store, 0)

	err := db.initState()
	if err != nil {
		return nil, err
	}
	if !db.RecoverTrie(stateRoot) {
		return nil, ErrRecoverRoot
	}

	db.cache = database.NewTransactionStore(db.trie)
	return db, nil
}

func (trieStore *TrieStore) initState() error {
	value, _ := trieStore.store.Get(trie.EmptyRoot[:])
	if value == nil {
		var err error
		trieStore.trie, err = trie.NewSecure(crypto.Hash{}, trieStore.trieDb)
		if err != nil {
			return err
		}

		trieStore.store.Put(trie.EmptyRoot[:], []byte{0})
	}

	return nil
}

func (trieStore *TrieStore) GetStorage(addr *crypto.CommonAddress) (*types.Storage, error) {
	storage := &types.Storage{}
	key := sha3.Keccak256([]byte(addressStorage + addr.Hex()))
	value, err := trieStore.Get(key)
	if err != nil {
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

func (trieStore *TrieStore) DeleteStorage(addr *crypto.CommonAddress) error {
	key := sha3.Keccak256([]byte(addressStorage + addr.Hex()))
	if trieStore.cache != nil {
		return trieStore.cache.Delete(key)
	} else {
		err := trieStore.trie.TryDelete(key)
		if err != nil {
			return err
		}
		_, err = trieStore.trie.Commit(nil)
		return err
	}
}

func (trieStore *TrieStore) PutStorage(addr *crypto.CommonAddress, storage *types.Storage) error {
	key := sha3.Keccak256([]byte(addressStorage + addr.Hex()))
	value, err := binary.Marshal(storage)
	if err != nil {
		return err
	}
	return trieStore.Put(key, value)
}

func (trieStore *TrieStore) Get(key []byte) ([]byte, error) {
	var value []byte
	var err error
	if trieStore.cache != nil {
		value, err = trieStore.cache.Get(key)
	} else {
		value, err = trieStore.trie.TryGet(key)
	}
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (trieStore *TrieStore) Put(key []byte, value []byte) error {
	if trieStore.cache != nil {
		return trieStore.cache.Put(key, value)
	} else {
		err := trieStore.trie.TryUpdate(key, value)
		if err != nil {
			return err
		}
		_, err = trieStore.trie.Commit(nil)
		return err
	}
}

func (trieStore *TrieStore) Delete(key []byte) error {
	return trieStore.store.Delete(key)
}

func (trieStore *TrieStore) GetBalance(addr *crypto.CommonAddress) *big.Int {
	storage, _ := trieStore.GetStorage(addr)

	if storage == nil {
		return new(big.Int)
	}
	return &storage.Balance
}

func (trieStore *TrieStore) PutBalance(addr *crypto.CommonAddress, balance *big.Int) error {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		storage = &types.Storage{}
	}
	storage.Balance = *balance
	return trieStore.PutStorage(addr, storage)
}

func (trieStore *TrieStore) GetNonce(addr *crypto.CommonAddress) uint64 {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		return 0
	}
	return storage.Nonce
}

func (trieStore *TrieStore) PutNonce(addr *crypto.CommonAddress, nonce uint64) error {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		storage = &types.Storage{}
	}
	storage.Nonce = nonce
	return trieStore.PutStorage(addr, storage)
}

func (trieStore *TrieStore) GetStorageAlias(addr *crypto.CommonAddress) string {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		return ""
	}
	return storage.Alias
}

func (trieStore *TrieStore) setStorageAlias(addr *crypto.CommonAddress, alias string) error {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		storage = &types.Storage{}
	}
	storage.Alias = alias
	return trieStore.PutStorage(addr, storage)
}

func (trieStore *TrieStore) AliasSet(addr *crypto.CommonAddress, alias string) (err error) {
	if alias != "" {
		//1 检查别名是否存在
		b := trieStore.AliasExist(alias)
		if b {
			return ErrUsedAlias
		}

		//2 存入以alias为key的k-v对
		err = trieStore.AliasPut(alias, addr.Bytes())
		if err != nil {
			return err
		}
	} else {
		return ErrInvalidateAlias
	}

	//put to stroage
	err = trieStore.setStorageAlias(addr, alias)
	if err != nil {
		return err
	}
	return nil
}

func (trieStore *TrieStore) AliasPut(alias string, value []byte) error {
	trieStore.cache.Put([]byte(aliasPrefix+alias), value)
	return nil
}

//alias为key的k-v
func (trieStore *TrieStore) AliasGet(alias string) (*crypto.CommonAddress, error) {
	buf, err := trieStore.store.Get([]byte(aliasPrefix + alias))
	if err != nil {
		return nil, err
	}
	addr := crypto.CommonAddress{}
	addr.SetBytes(buf)
	return &addr, nil
}

func (trieStore *TrieStore) AliasExist(alias string) bool {
	_, err := trieStore.store.Get([]byte(aliasPrefix + alias))
	if err != nil {
		return false
	}
	return true
}

func (trieStore *TrieStore) GetByteCode(addr *crypto.CommonAddress) []byte {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		return nil
	}
	return storage.ByteCode
}

func (trieStore *TrieStore) PutByteCode(addr *crypto.CommonAddress, byteCode []byte) error {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		storage = &types.Storage{}
	}
	storage.ByteCode = byteCode
	storage.CodeHash = crypto.GetByteCodeHash(byteCode)
	return trieStore.PutStorage(addr, storage)
}

func (trieStore *TrieStore) GetCodeHash(addr *crypto.CommonAddress) crypto.Hash {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		return crypto.Hash{}
	}
	return storage.CodeHash
}

func (trieStore *TrieStore) GetReputation(addr *crypto.CommonAddress) *big.Int {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		return big.NewInt(0)
	}
	return &storage.Reputation
}

func (trieStore *TrieStore) PutLogs(logs []*types.Log, txHash crypto.Hash) error {
	key := sha3.Keccak256([]byte("logs_" + txHash.String()))
	value, err := binary.Marshal(logs)
	if err != nil {
		return err
	}
	return trieStore.Put(key, value)
}

func (trieStore *TrieStore) AddBalance(addr *crypto.CommonAddress, amount *big.Int) error {
	balance := trieStore.GetBalance(addr)
	if balance == nil {
		balance = new(big.Int).SetInt64(0)
	}
	return trieStore.PutBalance(addr, new(big.Int).Add(balance, amount))
}

func (trieStore *TrieStore) SubBalance(addr *crypto.CommonAddress, amount *big.Int) error {
	balance := trieStore.GetBalance(addr)
	return trieStore.PutBalance(addr, new(big.Int).Sub(balance, amount))
}

func (trieStore *TrieStore) cacheToTrie() {
	if trieStore.cache != nil {
		trieStore.cache.Flush()
	}
}

func (trieStore *TrieStore) RevertState(shot *database.SnapShot) {
	trieStore.cache.RevertState(shot)
}

func (trieStore *TrieStore) CopyState() *database.SnapShot {
	return trieStore.cache.CopyState()
}

func (trieStore *TrieStore) GetStateRoot() []byte {
	trieStore.cacheToTrie()
	return trieStore.trie.Hash().Bytes()
}

func (trieStore *TrieStore) RecoverTrie(root []byte) bool {
	var err error
	trieStore.trie, err = trie.NewSecure(crypto.Bytes2Hash(root), trieStore.trieDb)
	if err != nil {
		return false
	}
	trieStore.cache = database.NewTransactionStore(trieStore.trie)
	return true
}

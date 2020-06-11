package store

import (
	"errors"
	"github.com/drep-project/DREP-Chain/common/trie"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"github.com/drep-project/DREP-Chain/types"
	"github.com/drep-project/binary"
	"math/big"
	"sync"
)

const (
	//AliasPrefix storage alias used prefix
	AliasPrefix = "alias"
	//AddressStorage Object stored with the address as the KEY
	AddressStorage = "AddressStorage"
)

var (
	//ErrRecoverRoot recover root err
	ErrRecoverRoot = errors.New("fail to recover root state")
	//ErrUsedAlias the alias has been used
	ErrUsedAlias = errors.New("the alias has been used")
	//ErrInvalidateAlias set null string as alias
	ErrInvalidateAlias = errors.New("set null string as alias")
)

type trieAccountStore struct {
	lock    sync.Mutex
	storeDB *StoreDB
}

func newTrieAccoutStore(store *StoreDB) *trieAccountStore {
	return &trieAccountStore{
		storeDB: store,
	}
}

func (trieStore *trieAccountStore) TrieDB() *trie.Database {
	return trieStore.storeDB.trieDb
}

func (trieStore *trieAccountStore) initState() error {
	value, _ := trieStore.storeDB.Get(trie.EmptyRoot[:])
	if value == nil {
		var err error
		trieStore.storeDB.initState()
		if err != nil {
			return err
		}

		trieStore.storeDB.Put(trie.EmptyRoot[:], []byte{0})
	}

	return nil
}

func (trieStore *trieAccountStore) GetStorage(addr *crypto.CommonAddress) (*types.Storage, error) {
	trieStore.lock.Lock()
	defer trieStore.lock.Unlock()

	storage := &types.Storage{}
	key := sha3.Keccak256([]byte(AddressStorage + addr.Hex()))
	value, err := trieStore.storeDB.Get(key)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}

	err = binary.Unmarshal(value, storage)
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (trieStore *trieAccountStore) DeleteStorage(addr *crypto.CommonAddress) error {
	trieStore.lock.Lock()
	defer trieStore.lock.Unlock()

	key := sha3.Keccak256([]byte(AddressStorage + addr.Hex()))

	return trieStore.storeDB.Delete(key)
}

func (trieStore *trieAccountStore) PutStorage(addr *crypto.CommonAddress, storage *types.Storage) error {
	trieStore.lock.Lock()
	defer trieStore.lock.Unlock()

	key := sha3.Keccak256([]byte(AddressStorage + addr.Hex()))
	value, err := binary.Marshal(storage)
	if err != nil {
		return err
	}
	return trieStore.storeDB.Put(key, value)
}

func (trieStore *trieAccountStore) GetBalance(addr *crypto.CommonAddress) *big.Int {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		return new(big.Int)
	}

	//Gets the currency in stakeStore
	return &storage.Balance
}

func (trieStore *trieAccountStore) PutBalance(addr *crypto.CommonAddress, balance *big.Int) error {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		storage = &types.Storage{}
	}
	storage.Balance = *balance
	return trieStore.PutStorage(addr, storage)
}

func (trieStore *trieAccountStore) GetNonce(addr *crypto.CommonAddress) uint64 {

	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		return 0
	}
	return storage.Nonce
}

func (trieStore *trieAccountStore) PutNonce(addr *crypto.CommonAddress, nonce uint64) error {

	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		storage = &types.Storage{}
	}
	storage.Nonce = nonce
	return trieStore.PutStorage(addr, storage)
}

func (trieStore *trieAccountStore) GetStorageAlias(addr *crypto.CommonAddress) string {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		return ""
	}
	return storage.Alias
}

func (trieStore *trieAccountStore) setStorageAlias(addr *crypto.CommonAddress, alias string) error {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		storage = &types.Storage{}
	}
	storage.Alias = alias
	return trieStore.PutStorage(addr, storage)
}

func (trieStore *trieAccountStore) AliasSet(addr *crypto.CommonAddress, alias string) (err error) {
	if alias != "" {
		//1 Check if the alias exists
		b := trieStore.AliasExist(alias)
		if b {
			return ErrUsedAlias
		}

		//2 Save the k-v pair with alias as the key
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

func (trieStore *trieAccountStore) AliasPut(alias string, value []byte) error {
	return trieStore.storeDB.Put([]byte(AliasPrefix+alias), value)
}

//AliasGet K--v for alias for key
func (trieStore *trieAccountStore) AliasGet(alias string) (*crypto.CommonAddress, error) {
	buf, err := trieStore.storeDB.Get([]byte(AliasPrefix + alias))
	if err != nil {
		return nil, err
	}
	addr := crypto.CommonAddress{}
	addr.SetBytes(buf)
	return &addr, nil
}

func (trieStore *trieAccountStore) AliasExist(alias string) bool {
	val, err := trieStore.storeDB.Get([]byte(AliasPrefix + alias))
	if val == nil || err != nil {
		return false
	}
	return true
}

func (trieStore *trieAccountStore) GetByteCode(addr *crypto.CommonAddress) []byte {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		return nil
	}
	return storage.ByteCode
}

func (trieStore *trieAccountStore) PutByteCode(addr *crypto.CommonAddress, byteCode []byte) error {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		storage = &types.Storage{}
	}
	storage.ByteCode = byteCode
	storage.CodeHash = crypto.GetByteCodeHash(byteCode)
	return trieStore.PutStorage(addr, storage)
}

func (trieStore *trieAccountStore) GetCodeHash(addr *crypto.CommonAddress) crypto.Hash {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		return crypto.Hash{}
	}
	return storage.CodeHash
}

func (trieStore *trieAccountStore) GetReputation(addr *crypto.CommonAddress) *big.Int {
	storage, _ := trieStore.GetStorage(addr)
	if storage == nil {
		return big.NewInt(0)
	}
	return &storage.Reputation
}

func (trieStore *trieAccountStore) PutLogs(logs []*types.Log, txHash crypto.Hash) error {
	key := sha3.Keccak256([]byte("logs_" + txHash.String()))
	value, err := binary.Marshal(logs)
	if err != nil {
		return err
	}
	return trieStore.storeDB.Put(key, value)
}

func (trieStore *trieAccountStore) AddBalance(addr *crypto.CommonAddress, amount *big.Int) error {
	balance := trieStore.GetBalance(addr)
	if balance == nil {
		balance = new(big.Int).SetInt64(0)
	}

	return trieStore.PutBalance(addr, new(big.Int).Add(balance, amount))
}

func (trieStore *trieAccountStore) SubBalance(addr *crypto.CommonAddress, amount *big.Int) error {
	balance := trieStore.GetBalance(addr)
	return trieStore.PutBalance(addr, new(big.Int).Sub(balance, amount))
}

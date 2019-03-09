package database

import (
    chainType "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/drep-chain/crypto"
    "github.com/drep-project/drep-chain/crypto/sha3"
    "encoding/json"
    "math/big"
)

type Cache struct {
    db     *Database
    memory map[string] []byte
    stores map[string] *chainType.Storage
    nodes  map[string] *Node
}

func NewCache(db *Database) *Cache {
    return &Cache {
        db:     db,
        memory: make(map[string] []byte),
        stores: make(map[string] *chainType.Storage),
        nodes:  make(map[string] *Node),
    }
}

func (cache *Cache) RootKey() []byte {
    return cache.db.trieRootKey
}

func (cache *Cache) GetNode(key []byte) (*Node, error) {
    k := bytes2Hex(key)
    node, ok := cache.nodes[k]
    if ok {
        return node, nil
    }
    value, err := cache.get(key)
    if err != nil {
        return nil, err
    }
    node = &Node{}
    err = json.Unmarshal(value, node)
    if err != nil {
        return nil, err
    }
    cache.nodes[k] = node
    return node, nil
}

func (cache *Cache) PutNode(key []byte, node *Node) error {
    value, err := json.Marshal(node)
    if err != nil {
        return err
    }
    cache.put(key, value)
    k := bytes2Hex(key)
    cache.nodes[k] = node
    return nil
}

func (cache *Cache) DelNode(key []byte) error {
    cache.del(key)
    k := bytes2Hex(key)
    cache.nodes[k] = nil
    return nil
}

func (cache *Cache) GetRootValue() []byte {
    return getRootValue(cache)
}

func (cache *Cache) get(key []byte) ([]byte, error) {
    k := bytes2Hex(key)
    value, ok := cache.memory[k]
    if ok {
        return value, nil
    }
    value, err := cache.db.get(key)
    if err != nil {
        return nil, err
    }
    cache.memory[k] = value
    return value, nil
}

func (cache *Cache) put(key, value []byte) {
    k := bytes2Hex(key)
    cache.memory[k] = value
}

func (cache *Cache) del(key []byte) {
    k := bytes2Hex(key)
    cache.memory[k] = nil
}

func (cache *Cache) commit() {
    for k, value := range cache.memory {
        key := hex2Bytes(k)
        if value != nil {
            cache.db.put(key, value, false)
        } else {
            cache.db.delete(key, false)
        }
    }
    cache.memory = make(map[string] []byte)
}

func (cache *Cache) discard() {
    cache.memory = make(map[string] []byte)
}

func (cache *Cache) getStorage(addr *crypto.CommonAddress) *chainType.Storage {
    key := sha3.Hash256([]byte("storage_" + addr.Hex()))
    k := bytes2Hex(key)
    storage, ok := cache.stores[k]
    if ok {
        return storage
    }
    storage = chainType.NewStorage()
    value, err := cache.get(key)
    if err == nil {
        json.Unmarshal(value, storage)
    }
    cache.stores[k] = storage
    return storage
}

func (cache *Cache) putStorage(addr *crypto.CommonAddress, storage *chainType.Storage) error {
    key := sha3.Hash256([]byte("storage_" + addr.Hex()))
    k := bytes2Hex(key)
    value, err := json.Marshal(storage)
    if err != nil {
        return err
    }
    cache.memory[k] = value
    cache.stores[k] = storage
    trieKey := commonKey2TrieKey(key)
    trieValue := commonValue2TrieValue(value)
    return updateTrie(cache, trieKey, trieValue)
}

func (cache *Cache) getBalance(addr *crypto.CommonAddress) *big.Int {
    storage := cache.getStorage(addr)
    return storage.Balance
}

func (cache *Cache) putBalance(addr *crypto.CommonAddress, balance *big.Int) error {
    storage := cache.getStorage(addr)
    storage.Balance = balance
    return cache.putStorage(addr, storage)
}

func (cache *Cache) addBalance(addr *crypto.CommonAddress, balance *big.Int) error {
    storage := cache.getStorage(addr)
    storage.Balance.Add(storage.Balance, balance)
    return cache.putStorage(addr, storage)
}

func (cache *Cache) subBalance(addr *crypto.CommonAddress, balance *big.Int) error {
    storage := cache.getStorage(addr)
    storage.Balance.Sub(storage.Balance, balance)
    return cache.putStorage(addr, storage)
}

func (cache *Cache) getNonce(addr *crypto.CommonAddress) int64 {
    storage := cache.getStorage(addr)
    return storage.Nonce
}

func (cache *Cache) putNonce(addr *crypto.CommonAddress, nonce int64) error {
    storage := cache.getStorage(addr)
    storage.Nonce = nonce
    return cache.putStorage(addr, storage)
}

func (cache *Cache) getReputation(addr *crypto.CommonAddress) *big.Int {
    storage := cache.getStorage(addr)
    return storage.Reputation
}

func (cache *Cache) putReputation(addr *crypto.CommonAddress, reputation *big.Int) error {
    storage := cache.getStorage(addr)
    storage.Reputation = reputation
    return cache.putStorage(addr, storage)
}

func (cache *Cache) getByteCode(addr *crypto.CommonAddress) []byte {
    storage := cache.getStorage(addr)
    return storage.ByteCode
}

func (cache *Cache) getCodeHash(addr *crypto.CommonAddress) crypto.Hash {
    storage := cache.getStorage(addr)
    return storage.CodeHash
}

func (cache *Cache) putByteCode(addr *crypto.CommonAddress, byteCode []byte) error {
    storage := cache.getStorage(addr)
    storage.ByteCode = byteCode
    storage.CodeHash = crypto.GetByteCodeHash(byteCode)
    return cache.putStorage(addr, storage)
}
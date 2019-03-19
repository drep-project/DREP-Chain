package database

import (
    "encoding/json"
    chainType "github.com/drep-project/drep-chain/chain/types"
    "github.com/syndtr/goleveldb/leveldb"
    "math/big"
    "github.com/drep-project/drep-chain/crypto"
    "github.com/drep-project/drep-chain/crypto/sha3"
    "encoding/hex"
)

type Database struct {
    db          *leveldb.DB
    trieRootKey []byte
}

func NewDatabase(dbPath string) (*Database, error) {
    ldb, err := leveldb.OpenFile(dbPath, nil)
    if err != nil {
        return nil, err
    }
    db := &Database{db: ldb}
    err = db.initState()
    if err != nil {
        return nil, err
    }
    return db, nil
}


func (db *Database) initState() error {
    db.trieRootKey = sha3.Hash256([]byte("state rootState"))
    rootState := &Node {
        Sequence: "",
        Value:    []byte{0},
        IsLeaf:   true,
    }
    value, err := json.Marshal(rootState)
    if err != nil {
        return err
    }
    return db.put(db.trieRootKey, value, false)
}

func (db *Database) rootKey() []byte {
    return db.trieRootKey
}

func (db *Database) get(key []byte) ([]byte, error) {
    return db.db.Get(key, nil)
}

func (db *Database) put(key []byte, value []byte, isJournal bool) error {
    err := db.db.Put(key, value, nil)
    if err != nil {
        return err
    }
    if isJournal {
        return db.addJournal("put", key, value)
    }
    return nil
}

func (db *Database) delete(key []byte, isJournal bool) error {
    err := db.db.Delete(key, nil)
    if err != nil {
        return err
    }
    if isJournal {
        return db.addJournal("del", key, nil)
    }
    return nil
}

func (db *Database) getStorage(addr *crypto.CommonAddress) *chainType.Storage {
    key := sha3.Hash256([]byte("storage_" + addr.Hex()))
    value, err := db.get(key)
    if err != nil {
        return chainType.NewStorage()
    }
    storage := chainType.NewStorage()
    json.Unmarshal(value, storage)
    return storage
}

func (db *Database) putStorage(addr *crypto.CommonAddress, storage *chainType.Storage) error {
    key := sha3.Hash256([]byte("storage_" + addr.Hex()))
    value, err := json.Marshal(storage)
    if err != nil {
        return err
    }
    return db.put(key, value, true)
}

func (db *Database) getBalance(addr *crypto.CommonAddress) *big.Int {
    storage := db.getStorage(addr)
    return storage.Balance
}

func (db *Database) putBalance(addr *crypto.CommonAddress, balance *big.Int) error {
    storage := db.getStorage(addr)
    storage.Balance = balance
    return db.putStorage(addr, storage)
}

func (db *Database) addBalance(addr *crypto.CommonAddress, balance *big.Int) error {
    storage := db.getStorage(addr)
    storage.Balance.Add(storage.Balance, balance)
    return db.putStorage(addr, storage)
}

func (db *Database) subBalance(addr *crypto.CommonAddress, balance *big.Int) error {
    storage := db.getStorage(addr)
    storage.Balance.Sub(storage.Balance, balance)
    return db.putStorage(addr, storage)
}

func (db *Database) getNonce(addr *crypto.CommonAddress) int64 {
    storage := db.getStorage(addr)
    return storage.Nonce
}

func (db *Database) putNonce(addr *crypto.CommonAddress, nonce int64) error {
    storage := db.getStorage(addr)
    storage.Nonce = nonce
    return db.putStorage(addr, storage)
}

func (db *Database) getReputation(addr *crypto.CommonAddress) *big.Int {
    storage := db.getStorage(addr)
    return storage.Reputation
}

func (db *Database) putReputation(addr *crypto.CommonAddress, reputation *big.Int) error {
    storage := db.getStorage(addr)
    storage.Reputation = reputation
    return db.putStorage(addr, storage)
}

func (db *Database) getByteCode(addr *crypto.CommonAddress) []byte {
    storage := db.getStorage(addr)
    return storage.ByteCode
}

func (db *Database) getCodeHash(addr *crypto.CommonAddress) crypto.Hash {
    storage := db.getStorage(addr)
    return storage.CodeHash
}

func (db *Database) putByteCode(addr *crypto.CommonAddress, byteCode []byte) error {
    storage := db.getStorage(addr)
    storage.ByteCode = byteCode
    storage.CodeHash = crypto.GetByteCodeHash(byteCode)
    return db.putStorage(addr, storage)
}

func (db *Database) getLogs(txHash []byte, ) []*chainType.Log {
    key := sha3.Hash256([]byte("logs_" + hex.EncodeToString(txHash)))
    value, err := db.get(key)
    if err != nil {
        return make([]*chainType.Log, 0)
    }
    var logs []*chainType.Log
    err = json.Unmarshal(value, &logs)
    if err != nil {
        return make([]*chainType.Log, 0)
    }
    return logs
}

func (db *Database) putLogs(logs []*chainType.Log, txHash []byte, ) error {
    key := sha3.Hash256([]byte("logs_" + hex.EncodeToString(txHash)))
    value, err := json.Marshal(logs)
    if err != nil {
        return err
    }
    return db.put(key, value, true)
}

func (db *Database) addLog(log *chainType.Log) error {
    logs := db.getLogs(log.TxHash)
    logs = append(logs, log)
    return db.putLogs(logs, log.TxHash)
}

func (db *Database) load(x *big.Int) []byte {
    value, _ := db.get(x.Bytes())
    return value
}

func (db *Database) store(x, y *big.Int) error {
    return db.put(x.Bytes(), y.Bytes(), true)
}

func (db *Database) getBlock(hash *crypto.Hash) (*chainType.Block, error) {
    key := append(BlockPrefix, hash[:]...)
    value, err := db.get(key)
    if err != nil {
        return nil, err
    }
    block := &chainType.Block{}
    json.Unmarshal(value, block)
    return block, nil
}

func (db *Database) putBlock(block *chainType.Block) error {
    hash := block.Header.Hash()
    key := append(BlockPrefix, hash[:]...)
    value, err := json.Marshal(block)
    if err != nil {
        return err
    }
    return db.put(key, value, false)
}
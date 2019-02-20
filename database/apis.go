package database

import (
    "strconv"
    "encoding/json"
    chainType "github.com/drep-project/drep-chain/chain/types"
    accountTypes "github.com/drep-project/drep-chain/accounts/types"
    "github.com/drep-project/drep-chain/common"
    "github.com/drep-project/drep-chain/crypto/sha3"
    "github.com/drep-project/drep-chain/crypto"
    "math/big"
    "errors"
    "encoding/hex"
)

func (database *DatabaseService) GetStateRoot() []byte {
    return db.getStateRoot()
}

func (database *DatabaseService) GetBlock(height int64) *chainType.Block {
    key := sha3.Hash256([]byte("block_" + strconv.FormatInt(height, 10)))
    value, err := db.get(key, false)
    if err != nil {
        return nil
    }
    block := &chainType.Block{}
    json.Unmarshal(value, block)
    return block
}

func (database *DatabaseService) PutBlock(block *chainType.Block) error {
    key := sha3.Hash256([]byte("block_" + strconv.FormatInt(block.Header.Height, 10)))
    value, err := json.Marshal(block)
    if err != nil {
        return err
    }
    return db.put(key, value, false)
}

func (database *DatabaseService) GetBlocksFrom(start, size int64) []*chainType.Block {
    var (
        currentBlock =&chainType.Block{}
        height = start
        blocks = make([]*chainType.Block, 0)
    )
    for currentBlock != nil && (height < start + size || size == -1)  {
        currentBlock = database.GetBlock(height)
        if currentBlock != nil {
            blocks = append(blocks, currentBlock)
        }
        height += 1
    }
    return blocks
}

func (database *DatabaseService) GetAllBlocks() []*chainType.Block {
    return database.GetBlocksFrom(int64(0), int64(-1))
}

func (database *DatabaseService) GetHighestBlock() *chainType.Block {
    maxHeight := database.GetMaxHeight()
    return database.GetBlock(maxHeight)
}

func (database *DatabaseService) GetMaxHeight() int64 {
    key := sha3.Hash256([]byte("max_height"))
    value, err := db.get(key, false)
    if err != nil {
        return -1
    } else {
        return new(big.Int).SetBytes(value).Int64()
    }
}

func (database *DatabaseService) PutMaxHeight(height int64) error {
    key := sha3.Hash256([]byte("max_height"))
    value := new(big.Int).SetInt64(height).Bytes()
    return db.put(key, value, false)
}

func (database *DatabaseService) GetPreviousBlockHash() []byte {
    key := sha3.Hash256([]byte("previous_hash"))
    value, _ := db.get(key, false)
    return value
}

func (database *DatabaseService) PutPreviousBlockHash(value []byte) error {
    key := sha3.Hash256([]byte("previous_hash"))
    return db.put(key, value, false)
}

func (database *DatabaseService) GetPreviousBlockTimestamp() int64 {
    key := sha3.Hash256([]byte("previous_hash"))
    value, err := db.get(key, false)
    if err != nil {
        return -1
    }
    return new(big.Int).SetBytes(value).Int64()
}

func (database *DatabaseService) PutPreviousBlockTimestamp(timestamp int64) error {
    key := sha3.Hash256([]byte("previous_hash"))
    value := new(big.Int).SetInt64(timestamp).Bytes()
    return db.put(key, value, false)
}

func (database *DatabaseService) GetStorage(addr crypto.CommonAddress, chainId common.ChainIdType, transactional bool) *accountTypes.Storage {
    if !transactional {
        return getStorage(addr, chainId)
    }
    if db.stores == nil {
        db.stores = make(map[string] *accountTypes.Storage)
    }
    key := sha3.Hash256([]byte("storage_" + addr.Hex() + chainId.Hex()))
    hk := bytes2Hex(key)
    storage, ok := db.stores[hk]
    if ok {
        return storage
    }
    storage = getStorage(addr, chainId)
    db.stores[hk] = storage
    return storage
}

func (database *DatabaseService) PutStorage(addr crypto.CommonAddress, chainId common.ChainIdType, storage *accountTypes.Storage, transactional bool) error {
    if !transactional {
        return putStorage(addr, chainId, storage)
    }
    if db.stores == nil {
        db.stores = make(map[string] *accountTypes.Storage)
    }
    key := sha3.Hash256([]byte("storage_" + addr.Hex() + chainId.Hex()))
    value, err := json.Marshal(storage)
    if err != nil {
        return err
    }
    err = db.put(key, value, true)
    if err != nil {
        return err
    }
    db.stores[bytes2Hex(key)] = storage
    return nil
}

func (database *DatabaseService) GetBalance(addr crypto.CommonAddress, chainId common.ChainIdType, transactional bool) *big.Int {
    storage := database.GetStorage(addr, chainId, transactional)
    if storage == nil {
        return new(big.Int)
    }
    return storage.Balance
}

func (database *DatabaseService) PutBalance(addr crypto.CommonAddress, chainId common.ChainIdType, balance *big.Int, transactional bool) error {
    storage := database.GetStorage(addr, chainId, transactional)
    if storage == nil {
        return errors.New("no account storage found")
    }
    storage.Balance = balance
    return database.PutStorage(addr, chainId, storage, transactional)
}

func (database *DatabaseService) GetNonce(addr crypto.CommonAddress, chainId common.ChainIdType, transactional bool) int64 {
    storage := database.GetStorage(addr, chainId, transactional)
    if storage == nil {
        return -1
    }
    return storage.Nonce
}

func (database *DatabaseService) PutNonce(addr crypto.CommonAddress, chainId common.ChainIdType, nonce int64, transactional bool) error {
    storage := database.GetStorage(addr, chainId, transactional)
    if storage == nil {
        return errors.New("no account storage found")
    }
    storage.Nonce = nonce
    return database.PutStorage(addr, chainId, storage, transactional)
}

func (database *DatabaseService) GetByteCode(addr crypto.CommonAddress, chainId common.ChainIdType, transactional bool) []byte {
    storage := database.GetStorage(addr, chainId, transactional)
    if storage == nil {
        return nil
    }
    return storage.ByteCode
}

func (database *DatabaseService) PutByteCode(addr crypto.CommonAddress, chainId common.ChainIdType, byteCode []byte, transactional bool) error {
    storage := database.GetStorage(addr, chainId, transactional)
    if storage == nil {
        return errors.New("no account storage found")
    }
    storage.ByteCode = byteCode
    storage.CodeHash = crypto.GetByteCodeHash(byteCode)
    return database.PutStorage(addr, chainId, storage, transactional)
}

func (database *DatabaseService) GetCodeHash(addr crypto.CommonAddress, chainId common.ChainIdType, transactional bool) crypto.Hash {
    storage := database.GetStorage(addr, chainId, transactional)
    if storage == nil {
        return crypto.Hash{}
    }
    return storage.CodeHash
}

func (database *DatabaseService) GetLogs(txHash []byte, chainId common.ChainIdType) []*chainType.Log {
    key := sha3.Hash256([]byte("logs_" + hex.EncodeToString(txHash) + chainId.Hex()))
    value, err := db.get(key, false)
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

func (database *DatabaseService) PutLogs(logs []*chainType.Log, txHash []byte, chainId common.ChainIdType) error {
    key := sha3.Hash256([]byte("logs_" + hex.EncodeToString(txHash) + chainId.Hex()))
    value, err := json.Marshal(logs)
    if err != nil {
        return err
    }
    return db.put(key, value, false)
}

func (database *DatabaseService) AddLog(log *chainType.Log) error {
    logs := database.GetLogs(log.TxHash, log.ChainId)
    logs = append(logs, log)
    return database.PutLogs(logs, log.TxHash, log.ChainId)
}


func (database *DatabaseService) Load(x *big.Int) []byte {
    value, _ := db.get(x.Bytes(), true)
    return value
}

func (database *DatabaseService) Store(x, y *big.Int) error {
    return db.put(x.Bytes(), y.Bytes(), true)
}

func (database *DatabaseService) BeginTransaction() {
    db.BeginTransaction()
}

func (database *DatabaseService) EndTransaction() {
    db.EndTransaction()
}

func (database *DatabaseService) Commit() {
    db.Commit()
}

func  (database *DatabaseService) Discard() {
    db.Discard()
}

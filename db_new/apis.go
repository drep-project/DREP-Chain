package db_new

import (
    "strconv"
    "encoding/json"
    "BlockChainTest/bean"
    "BlockChainTest/mycrypto"
    "math/big"
    "BlockChainTest/config"
    "BlockChainTest/accounts"
    "errors"
    "encoding/hex"
)

func GetBlock(height int64) *bean.Block {
    key := mycrypto.Hash256([]byte("block_" + strconv.FormatInt(height, 10)))
    value, err := db.get(key, false)
    if err != nil {
        return nil
    }
    block := &bean.Block{}
    json.Unmarshal(value, block)
    return block
}

func PutBlock(block *bean.Block) error {
    key := mycrypto.Hash256([]byte("block_" + strconv.FormatInt(block.Header.Height, 10)))
    value, err := json.Marshal(block)
    if err != nil {
        return err
    }
    return db.put(key, value, false)
}

func GetBlocksFrom(start, size int64) []*bean.Block {
    var (
        currentBlock =&bean.Block{}
        height = start
        blocks = make([]*bean.Block, 0)
    )
    for currentBlock != nil && (height < start + size || size == -1)  {
        currentBlock = GetBlock(height)
        if currentBlock != nil {
            blocks = append(blocks, currentBlock)
        }
        height += 1
    }
    return blocks
}

func GetAllBlocks() []*bean.Block {
    return GetBlocksFrom(int64(0), int64(-1))
}

func GetHighestBlock() *bean.Block {
    maxHeight := GetMaxHeight()
    return GetBlock(maxHeight)
}

func GetMaxHeight() int64 {
    key := mycrypto.Hash256([]byte("max_height"))
    value, err := db.get(key, false)
    if err != nil {
        return -1
    } else {
        return new(big.Int).SetBytes(value).Int64()
    }
}

func PutMaxHeight(height int64) error {
    key := mycrypto.Hash256([]byte("max_height"))
    value := new(big.Int).SetInt64(height).Bytes()
    return db.put(key, value, false)
}

func GetPreviousBlockHash() []byte {
    key := mycrypto.Hash256([]byte("previous_hash"))
    value, _ := db.get(key, false)
    return value
}

func PutPreviousBlockHash(value []byte) error {
    key := mycrypto.Hash256([]byte("previous_hash"))
    return db.put(key, value, false)
}

func GetPreviousBlockTimestamp() int64 {
    key := mycrypto.Hash256([]byte("previous_hash"))
    value, err := db.get(key, false)
    if err != nil {
        return -1
    }
    return new(big.Int).SetBytes(value).Int64()
}

func PutPreviousBlockTimestamp(timestamp int64) error {
    key := mycrypto.Hash256([]byte("previous_hash"))
    value := new(big.Int).SetInt64(timestamp).Bytes()
    return db.put(key, value, false)
}

func GetStorage(addr accounts.CommonAddress, chainId config.ChainIdType, transactional bool) *accounts.Storage {
    if !transactional {
        return getStorage(addr, chainId)
    }
    if db.stores == nil {
        db.stores = make(map[string] *accounts.Storage)
    }
    key := mycrypto.Hash256([]byte("storage_" + addr.Hex() + chainId.Hex()))
    hk := bytes2Hex(key)
    storage, ok := db.stores[hk]
    if ok {
        return storage
    }
    storage = getStorage(addr, chainId)
    db.stores[hk] = storage
    return storage
}

func PutStorage(addr accounts.CommonAddress, chainId config.ChainIdType, storage *accounts.Storage, transactional bool) error {
    if !transactional {
        return putStorage(addr, chainId, storage)
    }
    if db.stores == nil {
        db.stores = make(map[string] *accounts.Storage)
    }
    key := mycrypto.Hash256([]byte("storage_" + addr.Hex() + chainId.Hex()))
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

func GetBalance(addr accounts.CommonAddress, chainId config.ChainIdType, transactional bool) *big.Int {
    storage := GetStorage(addr, chainId, transactional)
    if storage == nil {
        return new(big.Int)
    }
    return storage.Balance
}

func PutBalance(addr accounts.CommonAddress, chainId config.ChainIdType, balance *big.Int, transactional bool) error {
    storage := GetStorage(addr, chainId, transactional)
    if storage == nil {
        return errors.New("no account storage found")
    }
    storage.Balance = balance
    return PutStorage(addr, chainId, storage, transactional)
}

func GetNonce(addr accounts.CommonAddress, chainId config.ChainIdType, transactional bool) int64 {
    storage := GetStorage(addr, chainId, transactional)
    if storage == nil {
        return -1
    }
    return storage.Nonce
}

func PutNonce(addr accounts.CommonAddress, chainId config.ChainIdType, nonce int64, transactional bool) error {
    storage := GetStorage(addr, chainId, transactional)
    if storage == nil {
        return errors.New("no account storage found")
    }
    storage.Nonce = nonce
    return PutStorage(addr, chainId, storage, transactional)
}

func GetByteCode(addr accounts.CommonAddress, chainId config.ChainIdType, transactional bool) []byte {
    storage := GetStorage(addr, chainId, transactional)
    if storage == nil {
        return nil
    }
    return storage.ByteCode
}

func PutByteCode(addr accounts.CommonAddress, chainId config.ChainIdType, byteCode []byte, transactional bool) error {
    storage := GetStorage(addr, chainId, transactional)
    if storage == nil {
        return errors.New("no account storage found")
    }
    storage.ByteCode = byteCode
    storage.CodeHash = accounts.GetByteCodeHash(byteCode)
    return PutStorage(addr, chainId, storage, transactional)
}

func GetCodeHash(addr accounts.CommonAddress, chainId config.ChainIdType, transactional bool) accounts.Hash {
    storage := GetStorage(addr, chainId, transactional)
    if storage == nil {
        return accounts.Hash{}
    }
    return storage.CodeHash
}

func GetLogs(txHash []byte, chainId config.ChainIdType) []*bean.Log {
    key := mycrypto.Hash256([]byte("logs_" + hex.EncodeToString(txHash) + chainId.Hex()))
    value, err := db.get(key, false)
    if err != nil {
        return make([]*bean.Log, 0)
    }
    var logs []*bean.Log
    err = json.Unmarshal(value, &logs)
    if err != nil {
        return make([]*bean.Log, 0)
    }
    return logs
}

func PutLogs(logs []*bean.Log, txHash []byte, chainId config.ChainIdType) error {
    key := mycrypto.Hash256([]byte("logs_" + hex.EncodeToString(txHash) + chainId.Hex()))
    value, err := json.Marshal(logs)
    if err != nil {
        return err
    }
    return db.put(key, value, false)
}

func AddLog(log *bean.Log) error {
    logs := GetLogs(log.TxHash, log.ChainId)
    logs = append(logs, log)
    return PutLogs(logs, log.TxHash, log.ChainId)
}
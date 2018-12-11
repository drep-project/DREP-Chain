package database

import (
    "math/big"
    "strconv"
    "encoding/hex"
    "encoding/json"

    "BlockChainTest/mycrypto"
    "BlockChainTest/bean"
    "BlockChainTest/config"
    "BlockChainTest/accounts"
    
)

type DataBaseAPI struct {
	
}



func(dataBaseAPI *DataBaseAPI) GetBlock(height int64) *bean.Block {
    key := mycrypto.Hash256([]byte("block_" + strconv.FormatInt(height, 10)))
    value := db.get(key)
    block, _ := bean.UnmarshalBlock(value)
    return block
}

func(dataBaseAPI *DataBaseAPI) GetBlocksFrom(start, size int64) []*bean.Block {
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

func(dataBaseAPI *DataBaseAPI) GetAllBlocks() []*bean.Block {
    return GetBlocksFrom(int64(0), int64(-1))
}

func(dataBaseAPI *DataBaseAPI) GetHighestBlock() *bean.Block {
    maxHeight := GetMaxHeight()
    return GetBlock(maxHeight)
}

//TODO cannot sync
func(dataBaseAPI *DataBaseAPI) PutBlock(block *bean.Block) error {
    key := mycrypto.Hash256([]byte("block_" + strconv.FormatInt(block.Header.Height, 10)))
    value, _ := bean.MarshalBlock(block)
    return db.put(key, value, config.GetConfig().ChainId)
}

func(dataBaseAPI *DataBaseAPI) GetMaxHeight() int64 {
    key := mycrypto.Hash256([]byte("max_height"))
    if value := db.get(key); value == nil {
        return -1
    } else {
        return new(big.Int).SetBytes(value).Int64()
    }
}

func(dataBaseAPI *DataBaseAPI) PutMaxHeight(height int64) error {
    key := mycrypto.Hash256([]byte("max_height"))
    value := new(big.Int).SetInt64(height).Bytes()
    err := db.put(key, value, config.GetConfig().ChainId)
    if err != nil {
        return err
    }
    return nil
}

func(dataBaseAPI *DataBaseAPI) GetStorageOutsideTransaction(addr accounts.CommonAddress, chainId int64) *accounts.Storage {
    key := mycrypto.Hash256([]byte("storage_" + addr.Hex() + strconv.FormatInt(chainId, 10)))
    value := db.get(key)
    storage := &accounts.Storage{}
    if value == nil {
        return storage
    }
    json.Unmarshal(value, storage)
    return storage
}

func(dataBaseAPI *DataBaseAPI) PutStorageOutsideTransaction(storage *accounts.Storage, addr accounts.CommonAddress, chainId int64) error {
    key := mycrypto.Hash256([]byte("storage_" + addr.Hex() + strconv.FormatInt(chainId, 10)))
    value, err := json.Marshal(storage)
    if err != nil {
        return err
    }
    return db.put(key, value, chainId)
}

func(dataBaseAPI *DataBaseAPI) GetStorageInsideTransaction(t *Transaction, addr accounts.CommonAddress, chainId int64) *accounts.Storage {
    key := mycrypto.Hash256([]byte("storage_" + addr.Hex() + strconv.FormatInt(chainId, 10)))
    value := t.Get(key)
    storage := &accounts.Storage{}
    if value == nil {
        return storage
    }
    json.Unmarshal(value, storage)
    return storage
}

func(dataBaseAPI *DataBaseAPI) PutStorageInsideTransaction(t *Transaction, storage *accounts.Storage, addr accounts.CommonAddress, chainId int64) error {
    key := mycrypto.Hash256([]byte("storage_" + addr.Hex() + strconv.FormatInt(chainId, 10)))
    value, err := json.Marshal(storage)
    if err != nil {
        return err
    }
    t.Put(key, value, chainId)
    return nil
}

func(dataBaseAPI *DataBaseAPI) GetMostRecentBlocks(n int64) []*bean.Block {
    height := GetMaxHeight()
    if height == -1 {
        return nil
    }
    return GetBlocksFrom(height - n, n)
}

func(dataBaseAPI *DataBaseAPI) GetBalanceOutsideTransaction(addr accounts.CommonAddress, chainId int64) *big.Int {
    storage := GetStorageOutsideTransaction(addr, chainId)
    if storage.Balance == nil {
        return new(big.Int)
    }
    return storage.Balance
}

func(dataBaseAPI *DataBaseAPI) PutBalanceOutSideTransaction(addr accounts.CommonAddress, chainId int64, balance *big.Int) error {
    storage := GetStorageOutsideTransaction(addr, chainId)
    storage.Balance = balance
    return PutStorageOutsideTransaction(storage, addr, chainId)
}

func(dataBaseAPI *DataBaseAPI) GetBalanceInsideTransaction(t *Transaction, addr accounts.CommonAddress, chainId int64) *big.Int {
    storage := GetStorageInsideTransaction(t, addr, chainId)
    if storage.Balance == nil {
        return new(big.Int)
    }
    return storage.Balance
}

func(dataBaseAPI *DataBaseAPI) PutBalanceInsideTransaction(t *Transaction, addr accounts.CommonAddress, chainId int64, balance *big.Int) error {
    storage := GetStorageInsideTransaction(t, addr, chainId)
    storage.Balance = balance
    return PutStorageInsideTransaction(t, storage, addr, chainId)
}

func(dataBaseAPI *DataBaseAPI) GetNonceOutsideTransaction(addr accounts.CommonAddress, chainId int64) int64 {
    storage := GetStorageOutsideTransaction(addr, chainId)
    return storage.Nonce
}

func(dataBaseAPI *DataBaseAPI) PutNonceOutsideTransaction(addr accounts.CommonAddress, chainId, nonce int64) error {
    storage := GetStorageOutsideTransaction(addr, chainId)
    storage.Nonce = nonce
    return PutStorageOutsideTransaction(storage, addr, chainId)
}

func(dataBaseAPI *DataBaseAPI) GetNonceInsideTransaction(t *Transaction, addr accounts.CommonAddress, chainId int64) int64 {
    storage := GetStorageInsideTransaction(t, addr, chainId)
    return storage.Nonce
}

func(dataBaseAPI *DataBaseAPI) PutNonceInsideTransaction(t *Transaction, addr accounts.CommonAddress, chainId, nonce int64) error {
    storage := GetStorageInsideTransaction(t, addr, chainId)
    storage.Nonce = nonce
    return PutStorageInsideTransaction(t, storage, addr, chainId)
}

func(dataBaseAPI *DataBaseAPI) GetByteCodeOutsideTransaction(addr accounts.CommonAddress, chainId int64) []byte {
    storage := GetStorageOutsideTransaction(addr, chainId)
    return storage.ByteCode
}

func(dataBaseAPI *DataBaseAPI) PutByteCodeOutsideTransaction(addr accounts.CommonAddress, chainId int64, byteCode []byte) error {
    storage := GetStorageOutsideTransaction(addr, chainId)
    storage.ByteCode = byteCode
    storage.CodeHash = accounts.GetByteCodeHash(byteCode)
    return PutStorageOutsideTransaction(storage, addr, chainId)
}

func(dataBaseAPI *DataBaseAPI) GetByteCodeInsideTransaction(t *Transaction, addr accounts.CommonAddress, chainId int64) []byte {
    storage := GetStorageInsideTransaction(t, addr, chainId)
    return storage.ByteCode
}

func(dataBaseAPI *DataBaseAPI) PutByteCodeInsideTransaction(t *Transaction, addr accounts.CommonAddress, chainId int64, byteCode []byte) error {
    storage := GetStorageInsideTransaction(t, addr, chainId)
    storage.ByteCode = byteCode
    storage.CodeHash = accounts.GetByteCodeHash(byteCode)
    return PutStorageInsideTransaction(t, storage, addr, chainId)
}

func(dataBaseAPI *DataBaseAPI) GetCodeHashOutsideTransaction(addr accounts.CommonAddress, chainId int64) accounts.Hash {
    storage := GetStorageOutsideTransaction(addr, chainId)
    return storage.CodeHash
}

func(dataBaseAPI *DataBaseAPI) GetCodeHashInsideTransaction(t *Transaction, addr accounts.CommonAddress, chainId int64) accounts.Hash {
    storage := GetStorageInsideTransaction(t, addr, chainId)
    return storage.CodeHash
}

func(dataBaseAPI *DataBaseAPI) GetLogsOutsideTransaction(txHash []byte, chainId int64) []*bean.Log {
    key := mycrypto.Hash256([]byte("logs_" + hex.EncodeToString(txHash) + strconv.FormatInt(chainId, 10)))
    value := db.get(key)
    if value == nil {
        return make([]*bean.Log, 0)
    }
    var logs []*bean.Log
    err := json.Unmarshal(value, logs)
    if err != nil {
        return make([]*bean.Log, 0)
    }
    return logs
}

func(dataBaseAPI *DataBaseAPI) GetLogsInsideTransaction(t *Transaction, txHash []byte, chainId int64) []*bean.Log {
    key := mycrypto.Hash256([]byte("logs_" + hex.EncodeToString(txHash) + strconv.FormatInt(chainId, 10)))
    value := t.Get(key)
    if value == nil {
        return make([]*bean.Log, 0)
    }
    var logs []*bean.Log
    err := json.Unmarshal(value, logs)
    if err != nil {
        return make([]*bean.Log, 0)
    }
    return logs
}

func(dataBaseAPI *DataBaseAPI) PutLogsInsideTransaction(t *Transaction, logs []*bean.Log, txHash []byte, chainId int64) error {
    key := mycrypto.Hash256([]byte("logs_" + hex.EncodeToString(txHash) + strconv.FormatInt(chainId, 10)))
    value, err := json.Marshal(logs)
    if err != nil {
        return err
    }
    t.Put(key, value, chainId)
    return nil
}

func(dataBaseAPI *DataBaseAPI) AddLogInsideTransaction(t *Transaction, log *bean.Log) error {
    logs := GetLogsInsideTransaction(t, log.TxHash, log.ChainId)
    logs = append(logs, log)
    return PutLogsInsideTransaction(t, logs, log.TxHash, log.ChainId)
}
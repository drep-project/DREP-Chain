package database

import (
    "math/big"
    "strconv"
    "encoding/hex"
    "encoding/json"

    "BlockChainTest/mycrypto"
    "BlockChainTest/bean"
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

func(dataBaseAPI *DataBaseAPI) GetMaxHeight() int64 {
    key := mycrypto.Hash256([]byte("max_height"))
    if value := db.get(key); value == nil {
        return -1
    } else {
        return new(big.Int).SetBytes(value).Int64()
    }
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

func(dataBaseAPI *DataBaseAPI) GetNonceOutsideTransaction(addr accounts.CommonAddress, chainId int64) int64 {
    storage := GetStorageOutsideTransaction(addr, chainId)
    return storage.Nonce
}

func(dataBaseAPI *DataBaseAPI) GetByteCodeOutsideTransaction(addr accounts.CommonAddress, chainId int64) []byte {
    storage := GetStorageOutsideTransaction(addr, chainId)
    return storage.ByteCode
}


func(dataBaseAPI *DataBaseAPI) GetByteCodeInsideTransaction(t *Transaction, addr accounts.CommonAddress, chainId int64) []byte {
    storage := GetStorageInsideTransaction(t, addr, chainId)
    return storage.ByteCode
}


func(dataBaseAPI *DataBaseAPI) GetCodeHashOutsideTransaction(addr accounts.CommonAddress, chainId int64) accounts.Hash {
    storage := GetStorageOutsideTransaction(addr, chainId)
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

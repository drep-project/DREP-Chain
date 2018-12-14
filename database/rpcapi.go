package database

import (
    "math/big"
    "BlockChainTest/bean"
    "BlockChainTest/accounts"

    "BlockChainTest/config"
)

type DataBaseAPI struct {}

func(dataBaseAPI *DataBaseAPI) GetBlock(height int64) *bean.Block {
    return GetBlock(height)
}

func(dataBaseAPI *DataBaseAPI) GetBlocksFrom(start, size int64) []*bean.Block {
    return GetBlocksFrom(start, size)
}

func(dataBaseAPI *DataBaseAPI) GetAllBlocks() []*bean.Block {
    return GetAllBlocks()
}

func(dataBaseAPI *DataBaseAPI) GetHighestBlock() *bean.Block {
    return GetHighestBlock()
}

func(dataBaseAPI *DataBaseAPI) GetMaxHeight() int64 {
    return GetMaxHeight()
}

func(dataBaseAPI *DataBaseAPI) GetMostRecentBlocks(n int64) []*bean.Block {
    return GetMostRecentBlocks(n)
}

func(dataBaseAPI *DataBaseAPI) GetBalance(addr accounts.CommonAddress, chainId config.ChainIdType) *big.Int {
    return GetBalance(addr, chainId)
}

func(dataBaseAPI *DataBaseAPI) GetNonce(addr accounts.CommonAddress, chainId config.ChainIdType) int64 {
    return GetNonce(addr, chainId)
}

func(dataBaseAPI *DataBaseAPI) GetByteCode(addr accounts.CommonAddress, chainId config.ChainIdType) []byte {
    return GetByteCode(addr, chainId)
}

func(dataBaseAPI *DataBaseAPI) GetCodeHash(addr accounts.CommonAddress, chainId config.ChainIdType) accounts.Hash {
    return GetCodeHash(addr, chainId)
}

func(dataBaseAPI *DataBaseAPI) GetLogs(txHash []byte, chainId config.ChainIdType) []*bean.Log {
    return GetLogs(txHash, chainId)
}

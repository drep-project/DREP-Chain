package database

import (
    "encoding/binary"
    "encoding/json"
    chainType "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/drep-chain/crypto"
    "github.com/syndtr/goleveldb/leveldb/util"
    "math/big"
)

var (
    MetaDataPrefix = []byte("metaData_")
    ChainStatePrefix = []byte("chainState_")
    BlockPrefix = []byte("block_")
    BlockNodePrefix = []byte("blockNode_")
)

func (database *DatabaseService) GetStateRoot() []byte {
    return database.cache.GetRootValue()
}

func (database *DatabaseService) PutBlock(block *chainType.Block) error {
    return database.db.putBlock(block)
}

func (database *DatabaseService) GetBlock(hash *crypto.Hash) (*chainType.Block, error) {
    return database.db.getBlock(hash)
}

func (database *DatabaseService) BlockIterator(handle func(*chainType.Block) error) error {
    iter := database.db.db.NewIterator(util.BytesPrefix(BlockPrefix), nil)
    defer iter.Release()
    var err error
    for iter.Next() {
        block := &chainType.Block{}
        err = json.Unmarshal(iter.Value(), block)
        if err != nil {
            break
        }
        err = handle(block)
        if err != nil {
            break
        }
    }
    if err != nil {
        return err
    }
    return nil
}

func (database *DatabaseService) PutBlockNode(blockNode *chainType.BlockNode) error {
    header := blockNode.Header()
    value, err := json.Marshal(header)
    if err != nil {
        return err
    }
    key := database.blockIndexKey(blockNode.Hash, blockNode.Height)
    value = append(value, byte(blockNode.Status))    //TODO just for now , when change binary serilize, should change a better one
    return database.db.put(key, value, false)
}

func (database *DatabaseService) blockIndexKey(blockHash *crypto.Hash, blockHeight int64) []byte {
    indexKey := make([]byte, len(BlockNodePrefix)+crypto.HashLength+8)
    copy(indexKey[0:len(BlockNodePrefix)], BlockNodePrefix[:])
    binary.BigEndian.PutUint64(indexKey[len(BlockNodePrefix):len(BlockNodePrefix)+8], uint64(blockHeight))
    copy(indexKey[len(BlockNodePrefix)+8:len(BlockNodePrefix)+40], blockHash[:])
    return indexKey
}

func (database *DatabaseService) GetBlockNode(hash *crypto.Hash, height int64) (*chainType.BlockHeader, chainType.BlockStatus, error) {
    key := database.blockIndexKey(hash, height)
    value, err := database.db.get(key)
    if err != nil {
        return nil, 0, err
    }
    blockHeader := &chainType.BlockHeader{}
    json.Unmarshal(value[0:len(value)-1], blockHeader)
    status :=  value[len(value)-1:len(value)][0]
    return blockHeader, chainType.BlockStatus(status), nil
}

func (database *DatabaseService) BlockNodeIterator(handle func(*chainType.BlockHeader, chainType.BlockStatus) error) error {
    iter := database.db.db.NewIterator(util.BytesPrefix(BlockNodePrefix), nil)
    defer iter.Release()
    var err error
    for iter.Next() {
        val := iter.Value()
        blockHeader := &chainType.BlockHeader{}
        err = json.Unmarshal(val[0:len(val)-1], blockHeader)
        if err != nil {
            break
        }
        err = handle(blockHeader, chainType.BlockStatus(val[len(val)-1:len(val)][0]))
        if err != nil {
            break
        }
    }
    if err != nil {
        return err
    }
    return nil
}

func (database *DatabaseService) PutChainState(chainState *chainType.BestState) error {
    key := ChainStatePrefix
    value, err := json.Marshal(chainState)
    if err != nil {
        return err
    }
    return database.db.put(key, value, false)
}

func (database *DatabaseService) GetChainState() *chainType.BestState {
    key := ChainStatePrefix
    value, err := database.db.get(key)
    if err != nil {
        return nil
    }
    state := &chainType.BestState{}
    json.Unmarshal(value, state)
    return state
}

func (database *DatabaseService) Rollback2Block(height int64) {
    database.db.Rollback2BlockHeight(height)
}

func (database *DatabaseService) RecordBlockJournal(height int64) {
    database.db.BlockHeight2JournalIndex(height)
}

func (database *DatabaseService) GetStorage(addr *crypto.CommonAddress) *chainType.Storage {
    return database.db.getStorage(addr)
}

func (database *DatabaseService) GetCachedStorage(addr *crypto.CommonAddress) *chainType.Storage {
    return database.cache.getStorage(addr)
}

func (database *DatabaseService) CacheStorage(addr *crypto.CommonAddress, storage *chainType.Storage) error {
    return database.cache.putStorage(addr, storage)
}

func (database *DatabaseService) GetBalance(addr *crypto.CommonAddress) *big.Int {
    return database.db.getBalance(addr)
}

func (database *DatabaseService) GetCachedBalance(addr *crypto.CommonAddress) *big.Int {
    return database.cache.getBalance(addr)
}

func (database *DatabaseService) CacheBalance(addr *crypto.CommonAddress, balance *big.Int) error {
    return database.cache.putBalance(addr, balance)
}

func (database *DatabaseService) AddBalance(addr *crypto.CommonAddress, balance *big.Int) error {
    return database.cache.addBalance(addr, balance)
}

func (database *DatabaseService) SubBalance(addr *crypto.CommonAddress, balance *big.Int) error {
    return database.cache.subBalance(addr, balance)
}

func (database *DatabaseService) GetNonce(addr *crypto.CommonAddress) int64 {
    return database.db.getNonce(addr)
}

func (database *DatabaseService) GetCachedNonce(addr *crypto.CommonAddress) int64 {
    return database.cache.getNonce(addr)
}

func (database *DatabaseService) CacheNonce(addr *crypto.CommonAddress, nonce int64) error {
    return database.cache.putNonce(addr, nonce)
}

func (database *DatabaseService) GetReputation(addr *crypto.CommonAddress) *big.Int {
    return database.db.getReputation(addr)
}

func (database *DatabaseService) GetCachedReputation(addr *crypto.CommonAddress) *big.Int {
    return database.cache.getReputation(addr)
}

func (database *DatabaseService) CacheReputation(addr *crypto.CommonAddress, reputation *big.Int) error {
    return database.cache.putReputation(addr, reputation);
}

func (database *DatabaseService) GetByteCode(addr *crypto.CommonAddress) []byte {
    return database.db.getByteCode(addr)
}

func (database *DatabaseService) GetCachedByteCode(addr *crypto.CommonAddress) []byte {
    return database.cache.getByteCode(addr)
}

func (database *DatabaseService) GetCodeHash(addr *crypto.CommonAddress) crypto.Hash {
    return database.db.getCodeHash(addr)
}

func (database *DatabaseService) GetCachedCodeHash(addr *crypto.CommonAddress) crypto.Hash {
    return database.cache.getCodeHash(addr)
}

func (database *DatabaseService) CacheByteCode(addr *crypto.CommonAddress, byteCode []byte) error {
    return database.cache.putByteCode(addr, byteCode)
}

func (database *DatabaseService) GetLogs(txHash []byte, ) []*chainType.Log {
    return database.db.getLogs(txHash)
}

func (database *DatabaseService) PutLogs(logs []*chainType.Log, txHash []byte, ) error {
    return database.db.putLogs(logs, txHash)
}

func (database *DatabaseService) AddLog(log *chainType.Log) error {
    return database.db.addLog(log)
}

func (database *DatabaseService) Load(x *big.Int) []byte {
    return database.db.load(x)
}

func (database *DatabaseService) Store(x, y *big.Int) error {
    return database.db.store(x, y)
}

func (database *DatabaseService) BeginTransaction() {
    database.lock.Lock()
    database.cache = NewCache(database.db)
}

func (database *DatabaseService) EndTransaction() {
    database.cache = nil
    database.lock.Unlock()
}

func (database *DatabaseService) Commit() {
    database.cache.commit()
    database.EndTransaction()
}

func  (database *DatabaseService) Discard() {
    database.cache.discard()
    database.EndTransaction()
}
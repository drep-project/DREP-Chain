package database

import (
    "encoding/hex"
    "errors"
    "github.com/drep-project/binary"
    chainType "github.com/drep-project/drep-chain/chain/types"
    "github.com/drep-project/drep-chain/crypto"
    "github.com/drep-project/drep-chain/crypto/sha3"
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
    return database.db.getStateRoot()
}

func (database *DatabaseService) PutBlock(block *chainType.Block) error {
    hash := block.Header.Hash()
    key := append(BlockPrefix, hash[:]...)
    value, err := binary.Marshal(block)
    if err != nil {
        return err
    }
    return database.db.db.Put(key, value, nil)
}

func (database *DatabaseService) GetBlock(hash *crypto.Hash) (*chainType.Block, error) {
    key := append(BlockPrefix, hash[:]...)
    value, err := database.db.get(key, false)
    if err != nil {
        return nil, err
    }
    block := &chainType.Block{}
    binary.Unmarshal(value, block)
    return block, nil
}

func (database *DatabaseService) BlockIterator(handle func(*chainType.Block) error) error {
    iter := database.db.db.NewIterator(util.BytesPrefix(BlockPrefix), nil)
    defer iter.Release()
    var err error
    for iter.Next() {
        block := &chainType.Block{}
        err = binary.Unmarshal(iter.Value(), block)
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
    value, err := binary.Marshal(header)
    if err != nil {
        return err
    }
    key := database.blockIndexKey(blockNode.Hash, blockNode.Height)
    value = append(value, byte(blockNode.Status))    //TODO just for now , when change binary serilize, should change a better one
    return database.db.db.Put(key, value, nil)
}

func (database *DatabaseService) blockIndexKey(blockHash *crypto.Hash, blockHeight uint64) []byte {
    indexKey := make([]byte, len(BlockNodePrefix)+crypto.HashLength+8)
    copy(indexKey[0:len(BlockNodePrefix)], BlockNodePrefix[:])
    binary.BigEndian.PutUint64(indexKey[len(BlockNodePrefix):len(BlockNodePrefix)+8], uint64(blockHeight))
    copy(indexKey[len(BlockNodePrefix)+8:len(BlockNodePrefix)+40], blockHash[:])
    return indexKey
}

func (database *DatabaseService) GetBlockNode(hash *crypto.Hash, height uint64) (*chainType.BlockHeader, chainType.BlockStatus, error) {
    key := database.blockIndexKey(hash, height)
    value, err := database.db.get(key, false)
    if err != nil {
        return nil, 0, err
    }
    blockHeader := &chainType.BlockHeader{}
    binary.Unmarshal(value[0:len(value)-1], blockHeader)
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
        err = binary.Unmarshal(val[0:len(val)-1], blockHeader)
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
    value, err := binary.Marshal(chainState)
    if err != nil {
        return err
    }
    return database.db.db.Put(key, value, nil)
}

func (database *DatabaseService) GetChainState() *chainType.BestState {
    key := ChainStatePrefix
    value, err := database.db.get(key, false)
    if err != nil {
        return nil
    }
    state := &chainType.BestState{}
    binary.Unmarshal(value, state)
    return state
}

func (database *DatabaseService) Rollback2Block(height uint64) {
     database.db.Rollback2Block(height)
}

func (database *DatabaseService) RecordBlockJournal(height uint64) {
    database.db.RecordBlockJournal(height)
}

func (database *DatabaseService) GetStorage(addr *crypto.CommonAddress, transactional bool) *chainType.Storage {
    if !transactional {
        return database.db.getStorage(addr)
    }

    if database.db.stores == nil {
       database.db.stores = make(map[string] *chainType.Storage)
    }

    key := sha3.Hash256([]byte("storage_" + addr.Hex()))
    hk := bytes2Hex(key)
    storage, ok := database.db.stores[hk]
    if ok {
        return storage
    }
    storage =  database.db.getStorage(addr)
    database.db.stores[hk] = storage
    return storage
}

func (database *DatabaseService) PutStorage(addr *crypto.CommonAddress, storage *chainType.Storage, transactional bool) error {
    if !transactional {
        return database.db.putStorage(addr, storage)
    }
    if database.db.stores == nil {
        database.db.stores = make(map[string] *chainType.Storage)
    }
    key := sha3.Hash256([]byte("storage_" + addr.Hex()))
    value, err := binary.Marshal(storage)
    if err != nil {
        return err
    }
    err = database.db.put(key, value, true)
    if err != nil {
        return err
    }
    database.db.stores[bytes2Hex(key)] = storage
    insert(database.db, bytes2Hex(key), database.db.root, sha3.Hash256(value))
    return nil
}

func (database *DatabaseService) GetBalance(addr *crypto.CommonAddress, transactional bool) *big.Int {
    storage := database.GetStorage(addr, transactional)

    if storage == nil {
        return new(big.Int)
    }
    return &storage.Balance
}

func (database *DatabaseService) PutBalance(addr *crypto.CommonAddress, balance *big.Int, transactional bool) error {
    storage := database.GetStorage(addr, transactional)
    if storage == nil {
        return errors.New("no account storage found")
    }
    storage.Balance = *balance
    return database.PutStorage(addr, storage, transactional)
}

func (database *DatabaseService) GetNonce(addr *crypto.CommonAddress, transactional bool) uint64 {
    storage := database.GetStorage(addr, transactional)
    if storage == nil {
        return 0
    }
    return storage.Nonce
}

func (database *DatabaseService) PutNonce(addr *crypto.CommonAddress, nonce uint64, transactional bool) error {
    storage := database.GetStorage(addr, transactional)
    if storage == nil {
        return errors.New("no account storage found")
    }
    storage.Nonce = nonce
    return database.PutStorage(addr, storage, transactional)
}

func (database *DatabaseService) GetByteCode(addr *crypto.CommonAddress, transactional bool) []byte {
    storage := database.GetStorage(addr, transactional)
    if storage == nil {
        return nil
    }
    return storage.ByteCode
}

func (database *DatabaseService) PutByteCode(addr *crypto.CommonAddress, byteCode []byte, transactional bool) error {
    storage := database.GetStorage(addr, transactional)
    if storage == nil {
        return errors.New("no account storage found")
    }
    storage.ByteCode = byteCode
    storage.CodeHash = crypto.GetByteCodeHash(byteCode)
    return database.PutStorage(addr, storage, transactional)
}

func (database *DatabaseService) GetCodeHash(addr *crypto.CommonAddress, transactional bool) crypto.Hash {
    storage := database.GetStorage(addr, transactional)
    if storage == nil {
        return crypto.Hash{}
    }
    return storage.CodeHash
}

func (database *DatabaseService) GetReputation(addr *crypto.CommonAddress, transactional bool) *big.Int {
    storage := database.GetStorage(addr, transactional)
    if storage == nil {
        return big.NewInt(0)
    }
    return &storage.Reputation
}

func (database *DatabaseService) GetLogs(txHash []byte, ) []*chainType.Log {
    key := sha3.Hash256([]byte("logs_" + hex.EncodeToString(txHash)))
    value, err := database.db.get(key, false)
    if err != nil {
        return make([]*chainType.Log, 0)
    }
    var logs []*chainType.Log
    err = binary.Unmarshal(value, &logs)
    if err != nil {
        return make([]*chainType.Log, 0)
    }
    return logs
}

func (database *DatabaseService) PutLogs(logs []*chainType.Log, txHash []byte, ) error {
    key := sha3.Hash256([]byte("logs_" + hex.EncodeToString(txHash)))
    value, err := binary.Marshal(logs)
    if err != nil {
        return err
    }
    return database.db.put(key, value, false)
}

func (database *DatabaseService) AddLog(log *chainType.Log) error {
    logs := database.GetLogs(log.TxHash)
    logs = append(logs, log)
    return database.PutLogs(logs, log.TxHash)
}


func (database *DatabaseService) Load(x *big.Int) []byte {
    value, _ := database.db.get(x.Bytes(), true)
    return value
}

func (database *DatabaseService) Store(x, y *big.Int) error {
    return database.db.put(x.Bytes(), y.Bytes(), true)
}

func  (database *DatabaseService)Transaction(fn func() error) error{
    isSuccess := false
    database.BeginTransaction()
    defer func() {
        if isSuccess {
            database.Commit()
        }else{
            database.Discard()
        }
    }()
    err := fn()
    if err == nil {
        isSuccess = true
    }
    return err
}
func (database *DatabaseService) BeginTransaction() {
    database.db.BeginTransaction()
}

func (database *DatabaseService) EndTransaction() {
    database.db.EndTransaction()
}

func (database *DatabaseService) Commit() {
    database.db.Commit()
}

func  (database *DatabaseService) Discard() {
    database.db.Discard()
}

func (database *DatabaseService)AddBalance(addr *crypto.CommonAddress, amount *big.Int, transactional bool) {
    balance := database.GetBalance(addr, transactional)
    //text, _ := addr.MarshalText()
    //x := string(text)
    //fmt.Println("0x" + x)
    if balance == nil {
        balance = new(big.Int).SetInt64(0)
    }
    database.PutBalance(addr, new(big.Int).Add(balance, amount), transactional)
    return
}
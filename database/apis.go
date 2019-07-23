package database

import (
	"math/big"

	"github.com/drep-project/binary"
	chainType "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
)

var (
	MetaDataPrefix   = []byte("metaData_")
	ChainStatePrefix = []byte("chainState_")
	BlockPrefix      = []byte("block_")
	BlockNodePrefix  = []byte("blockNode_")
)

func (database *DatabaseService) GetStateRoot() []byte {
	return database.db.GetStateRoot()
}

func (database *DatabaseService) PutBlock(block *chainType.Block) error {
	hash := block.Header.Hash()
	key := append(BlockPrefix, hash[:]...)
	value, err := binary.Marshal(block)
	if err != nil {
		return err
	}

	return database.db.Put(key, value)
}

func (database *DatabaseService) GetBlock(hash *crypto.Hash) (*chainType.Block, error) {
	key := append(BlockPrefix, hash[:]...)
	value, err := database.db.Get(key)
	if err != nil {
		return nil, err
	}
	block := &chainType.Block{}
	err = binary.Unmarshal(value, block)
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (database *DatabaseService) HasBlock(hash *crypto.Hash) bool {
	key := append(BlockPrefix, hash[:]...)
	_, err := database.db.Get(key)
	return err == nil
}

func (database *DatabaseService) BlockIterator(handle func(*chainType.Block) error) error {
	iter := database.db.diskDb.NewIteratorWithPrefix(BlockPrefix)
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
	value = append(value, byte(blockNode.Status)) //TODO just for now , when change binary serilize, should change a better one
	return database.db.diskDb.Put(key, value)
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
	value, err := database.db.Get(key)
	if err != nil {
		return nil, 0, err
	}
	blockHeader := &chainType.BlockHeader{}
	binary.Unmarshal(value[0:len(value)-1], blockHeader)
	status := value[len(value)-1 : len(value)][0]
	return blockHeader, chainType.BlockStatus(status), nil
}

func (database *DatabaseService) BlockNodeCount() int64 {
	count := int64(64)
	iter := database.db.diskDb.NewIteratorWithPrefix(BlockNodePrefix)
	defer iter.Release()
	for iter.Next() {
		count++
	}
	return count
}

func (database *DatabaseService) BlockNodeIterator(handle func(*chainType.BlockHeader, chainType.BlockStatus) error) error {
	iter := database.db.diskDb.NewIteratorWithPrefix(BlockNodePrefix)
	defer iter.Release()
	var err error
	for iter.Next() {
		val := iter.Value()
		blockHeader := &chainType.BlockHeader{}
		err = binary.Unmarshal(val[0:len(val)-1], blockHeader)
		if err != nil {
			break
		}
		err = handle(blockHeader, chainType.BlockStatus(val[len(val)-1 : len(val)][0]))
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
	return database.db.PutChainState(chainState)
}

func (database *DatabaseService) GetChainState() *chainType.BestState {
	return database.db.GetChainState()
}

//返回回滚的操作数目
func (database *DatabaseService) Rollback2Block(height uint64) (error, int64) {
	return database.db.Rollback2Block(height)
}

func (database *DatabaseService) RecordBlockJournal(height uint64) {
	database.db.SetBlockJournal(height)
}

func (database *DatabaseService) GetBlockJournal() uint64 {
	return database.db.GetBlockJournal()
}

func (database *DatabaseService) GetStorage(addr *crypto.CommonAddress) *chainType.Storage {
	return database.db.GetStorage(addr)
}

func (database *DatabaseService) PutStorage(addr *crypto.CommonAddress, storage *chainType.Storage) error {
	return database.db.PutStorage(addr, storage)
}

func (database *DatabaseService) GetBalance(addr *crypto.CommonAddress) *big.Int {
	return database.db.GetBalance(addr)
}

func (database *DatabaseService) PutBalance(addr *crypto.CommonAddress, balance *big.Int) error {
	return database.db.PutBalance(addr, balance)
}

func (database *DatabaseService) GetNonce(addr *crypto.CommonAddress) uint64 {
	return database.db.GetNonce(addr)
}

func (database *DatabaseService) PutNonce(addr *crypto.CommonAddress, nonce uint64) error {
	return database.db.PutNonce(addr, nonce)
}

func (database *DatabaseService) GetStorageAlias(addr *crypto.CommonAddress) string {
	return database.db.GetStorageAlias(addr)
}

func (database *DatabaseService) AliasSet(addr *crypto.CommonAddress, alias string) (err error) {
	return database.db.AliasSet(addr, alias)
}

func (database *DatabaseService) AliasGet(alias string) *crypto.CommonAddress {
	return database.db.AliasGet(alias)
}

func (database *DatabaseService) AliasExist(alias string) bool {
	return database.db.AliasExist(alias)
}

func (database *DatabaseService) GetByteCode(addr *crypto.CommonAddress) []byte {
	return database.db.GetByteCode(addr)
}

func (database *DatabaseService) PutByteCode(addr *crypto.CommonAddress, byteCode []byte) error {
	return database.db.PutByteCode(addr, byteCode)
}

func (database *DatabaseService) GetCodeHash(addr *crypto.CommonAddress) crypto.Hash {
	return database.db.GetCodeHash(addr)
}

func (database *DatabaseService) GetReputation(addr *crypto.CommonAddress) *big.Int {
	return database.db.GetReputation(addr)
}

func (database *DatabaseService) GetLogs(txHash crypto.Hash) []*chainType.Log {
	return database.db.GetLogs(txHash)
}

func (database *DatabaseService) PutLogs(logs []*chainType.Log, txHash crypto.Hash) error {
	return database.db.PutLogs(logs, txHash)
}

func (database *DatabaseService) AddLog(log *chainType.Log) error {
	return database.db.AddLog(log)
}

func (database *DatabaseService) PutReceipt(txHash crypto.Hash, receipt *chainType.Receipt) error {
	return database.db.PutReceipt(txHash, receipt)
}

func (database *DatabaseService) GetReceipt(txHash crypto.Hash) *chainType.Receipt {
	return database.db.GetReceipt(txHash)
}

func (database *DatabaseService) PutReceipts(blockHash crypto.Hash, receipts []*chainType.Receipt) error {
	return database.db.PutReceipts(blockHash, receipts)
}

func (database *DatabaseService) GetReceipts(blockHash crypto.Hash) []*chainType.Receipt {
	return database.db.GetReceipts(blockHash)
}

func (database *DatabaseService) Load(x *big.Int) []byte {
	return database.db.Load(x)
}

func (database *DatabaseService) Store(x, y *big.Int) error {
	return database.db.Store(x, y)
}

func (database *DatabaseService) AddBalance(addr *crypto.CommonAddress, amount *big.Int) error {
	return database.db.AddBalance(addr, amount)
}

func (database *DatabaseService) BeginTransaction(storeToDB bool) *Database {
	return database.db.BeginTransaction(storeToDB)
}

func (database *DatabaseService) Commit(needLog bool) {
	database.db.Commit(needLog)
}

func (database *DatabaseService) Discard() {
	database.db.Discard()
}

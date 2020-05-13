package chain

import (
	"fmt"
	"github.com/drep-project/binary"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"github.com/drep-project/DREP-Chain/database/dbinterface"
	"github.com/drep-project/DREP-Chain/types"
)

var (
	MetaDataPrefix   = []byte("metaData_")
	ChainStatePrefix = []byte("chainState_")
	BlockPrefix      = []byte("block_")
	BlockNodePrefix  = []byte("blockNode_")
)

type ChainStore struct {
	dbinterface.KeyValueStore
}

func (chainStore *ChainStore) PutReceipt(txHash crypto.Hash, receipt *types.Receipt) error {
	key := sha3.Keccak256([]byte("receipt_" + txHash.String()))
	value, err := binary.Marshal(receipt)
	//fmt.Println("err11: ", err)
	if err != nil {
		return err
	}
	return chainStore.Put(key, value)
}

func (chainStore *ChainStore) GetReceipt(txHash crypto.Hash) *types.Receipt {
	key := sha3.Keccak256([]byte("receipt_" + txHash.String()))
	value, err := chainStore.Get(key)
	//fmt.Println("err12: ", err)
	//fmt.Println("val: ", value)
	if err != nil {
		return nil
	}
	receipt := &types.Receipt{}
	err = binary.Unmarshal(value, receipt)
	fmt.Println("err13: ", err)
	if err != nil {
		return nil
	}
	return receipt
}

func (chainStore *ChainStore) PutReceipts(blockHash crypto.Hash, receipts []*types.Receipt) error {
	key := sha3.Keccak256([]byte("receipts_" + blockHash.String()))
	value, err := binary.Marshal(receipts)
	if err != nil {
		return err
	}
	return chainStore.Put(key, value)
}

func (chainStore *ChainStore) GetReceipts(blockHash crypto.Hash) []*types.Receipt {
	key := sha3.Keccak256([]byte("receipts_" + blockHash.String()))
	value, err := chainStore.Get(key)
	if err != nil {
		return make([]*types.Receipt, 0)
	}
	var receipts []*types.Receipt
	err = binary.Unmarshal(value, &receipts)
	if err != nil {
		return make([]*types.Receipt, 0)
	}
	return receipts
}

func (chainStore *ChainStore) DeleteReceipts(blockHash crypto.Hash) error {
	key := sha3.Keccak256([]byte("receipts_" + blockHash.String()))
	return chainStore.Delete(key)
}

func (chainStore *ChainStore) PutBlock(block *types.Block) error {
	hash := block.Header.Hash()
	key := append(BlockPrefix, hash[:]...)
	value, err := binary.Marshal(block)
	if err != nil {
		return err
	}

	return chainStore.Put(key, value)
}

func (chainStore *ChainStore) GetBlockHeader(hash *crypto.Hash) (*types.BlockHeader, error) {
	key := append(BlockPrefix, hash[:]...)
	value, err := chainStore.Get(key)
	if err != nil {
		return nil, err
	}
	block := &types.Block{}
	err = binary.Unmarshal(value, block)
	if err != nil {
		return nil, err
	}

	return block.Header, nil
}

func (chainStore *ChainStore) GetBlock(hash *crypto.Hash) (*types.Block, error) {
	key := append(BlockPrefix, hash[:]...)
	value, err := chainStore.Get(key)
	if err != nil {
		return nil, err
	}
	block := &types.Block{}
	err = binary.Unmarshal(value, block)
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (chainStore *ChainStore) HasBlock(hash *crypto.Hash) bool {
	key := append(BlockPrefix, hash[:]...)
	_, err := chainStore.Get(key)
	return err == nil
}

func (chainStore *ChainStore) PutBlockNode(blockNode *types.BlockNode) error {
	header := blockNode.Header()
	value, err := binary.Marshal(header)
	if err != nil {
		return err
	}
	key := chainStore.blockIndexKey(blockNode.Hash, blockNode.Height)
	value = append(value, byte(blockNode.Status))
	return chainStore.Put(key, value)
}

func (chainStore *ChainStore) blockIndexKey(blockHash *crypto.Hash, blockHeight uint64) []byte {
	indexKey := make([]byte, len(BlockNodePrefix)+crypto.HashLength+8)
	copy(indexKey[0:len(BlockNodePrefix)], BlockNodePrefix[:])
	binary.BigEndian.PutUint64(indexKey[len(BlockNodePrefix):len(BlockNodePrefix)+8], uint64(blockHeight))
	copy(indexKey[len(BlockNodePrefix)+8:len(BlockNodePrefix)+40], blockHash[:])
	return indexKey
}

func (chainStore *ChainStore) GetBlockNode(hash *crypto.Hash, height uint64) (*types.BlockHeader, types.BlockStatus, error) {
	key := chainStore.blockIndexKey(hash, height)
	value, err := chainStore.Get(key)
	if err != nil {
		return nil, 0, err
	}
	blockHeader := &types.BlockHeader{}
	binary.Unmarshal(value[0:len(value)-1], blockHeader)
	status := value[len(value)-1 : len(value)][0]
	return blockHeader, types.BlockStatus(status), nil
}

func (chainStore *ChainStore) BlockNodeCount() int64 {
	count := int64(64)
	iter := chainStore.NewIteratorWithPrefix(BlockNodePrefix)
	defer iter.Release()
	for iter.Next() {
		count++
	}
	return count
}

func (chainStore *ChainStore) BlockNodeIterator(handle func(*types.BlockHeader, types.BlockStatus) error) error {
	iter := chainStore.NewIteratorWithPrefix(BlockNodePrefix)
	defer iter.Release()
	var err error
	for iter.Next() {
		val := iter.Value()
		blockHeader := &types.BlockHeader{}
		err = binary.Unmarshal(val[0:len(val)-1], blockHeader)
		if err != nil {
			break
		}
		err = handle(blockHeader, types.BlockStatus(val[len(val)-1 : len(val)][0]))
		if err != nil {
			break
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func (chainStore *ChainStore) RollBack(height uint64, hash *crypto.Hash) (error, int64) {
	var err error

	//删除blocknode
	func() {
		key := chainStore.blockIndexKey(hash, height)
		err = chainStore.Delete(key)
	}()

	if err != nil {
		return err, 0
	}

	//删除block
	func() {
		key := append(BlockPrefix, hash[:]...)
		err = chainStore.Delete(key)
	}()

	if err != nil {
		return err, 0
	}

	return nil, 0
}

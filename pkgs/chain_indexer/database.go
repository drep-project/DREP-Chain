package chain_indexer

import (
	bin "encoding/binary"
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/crypto"
)

var (
	indexerPrefix = "ci_"
	bloomPrefix   = []byte("B")
)

func (chainIndexer *ChainIndexerService) getStoredSections() uint64 {
	var storedSections uint64
	value, err := chainIndexer.DatabaseService.Get([]byte(indexerPrefix + "count"))
	if err != nil {
		return storedSections
	}

	binary.Unmarshal(value, &storedSections)
	return storedSections
}

func (chainIndexer *ChainIndexerService) setStoredSections(storedSections uint64) error {
	value, err := binary.Marshal(storedSections)
	if err != nil {
		return err
	}
	return chainIndexer.DatabaseService.Put([]byte(indexerPrefix+"count"), value)
}

// GetSectionHead 从数据库中获取已处理section的最后一个块哈希
func (chainIndexer *ChainIndexerService) getSectionHead(section uint64) crypto.Hash {
	var data [8]byte
	bin.BigEndian.PutUint64(data[:], section)
	key := append([]byte(indexerPrefix+"shead"), data[:]...)

	var sectionHead crypto.Hash
	value, _ := chainIndexer.DatabaseService.Get(key)
	if len(value) == 0 {
		return sectionHead
	}

	binary.Unmarshal(value, sectionHead)
	return sectionHead
}

// SetSectionHead 将已处理section的最后一个块哈希写入数据库
func (chainIndexer *ChainIndexerService) setSectionHead(section uint64, hash crypto.Hash) error {
	var data [8]byte
	bin.BigEndian.PutUint64(data[:], section)
	key := append([]byte(indexerPrefix+"shead"), data[:]...)

	value, err := binary.Marshal(hash)
	if err != nil {
		return err
	}

	err = chainIndexer.DatabaseService.Put(key, value)
	if err != nil {
		return err
	}

	return nil
}

// DeleteSectionHead 将已处理section的最后一个块哈希从数据库中删除
func (chainIndexer *ChainIndexerService) deleteSectionHead(section uint64) error {
	var data [8]byte
	binary.BigEndian.PutUint64(data[:], section)
	key := append([]byte(indexerPrefix+"shead"), data[:]...)

	return chainIndexer.DatabaseService.Delete(key)
}

//bloomBitsKey = bloomBitsPrefix + bit (uint16 big endian) + section (uint64 big endian) + hash
func bloomBitsKey(bit uint, section uint64, hash crypto.Hash) []byte {
	key := append(append(bloomPrefix, make([]byte, 10)...), hash.Bytes()...)

	binary.BigEndian.PutUint16(key[1:], uint16(bit))
	binary.BigEndian.PutUint64(key[3:], section)

	return key
}

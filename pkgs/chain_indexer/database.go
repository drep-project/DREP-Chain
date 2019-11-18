package chain_indexer

import (
	bin "encoding/binary"
	"github.com/drep-project/binary"
	"github.com/drep-project/DREP-Chain/crypto"
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

// setValidStoredSections writes the number of valid sections to the index database
func (chainIndexer *ChainIndexerService) setValidStoredSections(sections uint64) {
	// Set the current number of valid sections in the database
	chainIndexer.setStoredSections(sections)

	// Remove any reorged sections, caching the valids in the mean time
	for chainIndexer.storedSections > sections {
		chainIndexer.storedSections--
		chainIndexer.deleteSectionHead(chainIndexer.storedSections)
	}
	chainIndexer.storedSections = sections // needed if new > old
}

// Sections returns the number of processed sections maintained by the indexer
// and also the information about the last header indexed for potential canonical
// verifications.
func (chainIndexer *ChainIndexerService) Sections() (uint64, uint64, crypto.Hash) {
	chainIndexer.lock.Lock()
	defer chainIndexer.lock.Unlock()

	chainIndexer.verifyLastHead()
	return chainIndexer.storedSections, chainIndexer.storedSections*chainIndexer.Config.SectionSize - 1, chainIndexer.getSectionHead(chainIndexer.storedSections - 1)
}

func (chainIndexer *ChainIndexerService) BloomStatus() (uint64, uint64) {
	sections, _, _ := chainIndexer.Sections()
	return chainIndexer.Config.SectionSize, sections
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
	sectionHead.SetBytes(value)
	return sectionHead
}

// SetSectionHead 将已处理section的最后一个块哈希写入数据库
func (chainIndexer *ChainIndexerService) setSectionHead(section uint64, hash crypto.Hash) error {
	var data [8]byte
	bin.BigEndian.PutUint64(data[:], section)
	key := append([]byte(indexerPrefix+"shead"), data[:]...)
	err := chainIndexer.DatabaseService.Put(key, hash.Bytes())
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

// ReadBloomBits retrieves the compressed bloom bit vector belonging to the given
// section and bit index from the.
func (chainIndexer *ChainIndexerService) ReadBloomBits(bit uint, section uint64, head crypto.Hash) ([]byte, error) {
	return chainIndexer.DatabaseService.Get(bloomBitsKey(bit, section, head))
}

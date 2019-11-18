package chain_indexer

import (
	bin "encoding/binary"
	"github.com/drep-project/binary"
	"github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/database/dbinterface"
	"github.com/drep-project/DREP-Chain/types"
)

var (
	indexerPrefix = "ci_"
	bloomPrefix   = []byte("B")
)

type ChainIndexerStore struct {
	dbinterface.KeyValueStore
	chainStore *chain.ChainStore
}

func NewChainIndexerStore(store dbinterface.KeyValueStore) *ChainIndexerStore {
	return &ChainIndexerStore{
		store,
		&chain.ChainStore{store},
	}
}

func (chainIndexerStore *ChainIndexerStore) getStoredSections() uint64 {
	var storedSections uint64
	value, err := chainIndexerStore.Get([]byte(indexerPrefix + "count"))
	if err != nil {
		return storedSections
	}

	binary.Unmarshal(value, &storedSections)
	return storedSections
}

func (chainIndexerStore *ChainIndexerStore) setStoredSections(storedSections uint64) error {
	value, err := binary.Marshal(storedSections)
	if err != nil {
		return err
	}
	return chainIndexerStore.Put([]byte(indexerPrefix+"count"), value)
}

// GetSectionHead 从数据库中获取已处理section的最后一个块哈希
func (chainIndexerStore *ChainIndexerStore) getSectionHead(section uint64) crypto.Hash {
	var data [8]byte
	bin.BigEndian.PutUint64(data[:], section)
	key := append([]byte(indexerPrefix+"shead"), data[:]...)

	var sectionHead crypto.Hash
	value, _ := chainIndexerStore.Get(key)
	if len(value) == 0 {
		return sectionHead
	}
	sectionHead.SetBytes(value)
	return sectionHead
}

// SetSectionHead 将已处理section的最后一个块哈希写入数据库
func (chainIndexerStore *ChainIndexerStore) setSectionHead(section uint64, hash crypto.Hash) error {
	var data [8]byte
	bin.BigEndian.PutUint64(data[:], section)
	key := append([]byte(indexerPrefix+"shead"), data[:]...)
	err := chainIndexerStore.Put(key, hash.Bytes())
	if err != nil {
		return err
	}
	return nil
}

// DeleteSectionHead 将已处理section的最后一个块哈希从数据库中删除
func (chainIndexerStore *ChainIndexerStore) deleteSectionHead(section uint64) error {
	var data [8]byte
	binary.BigEndian.PutUint64(data[:], section)
	key := append([]byte(indexerPrefix+"shead"), data[:]...)

	return chainIndexerStore.Delete(key)
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
func (chainIndexerStore *ChainIndexerStore) ReadBloomBits(bit uint, section uint64, head crypto.Hash) ([]byte, error) {
	return chainIndexerStore.Get(bloomBitsKey(bit, section, head))
}

func (chainIndexerStore *ChainIndexerStore) FindCommonAncestor(a, b *types.BlockHeader) *types.BlockHeader {
	for bn := b.Height; a.Height > bn; {
		a, _ := chainIndexerStore.chainStore.GetBlockHeader(&a.PreviousHash)
		if a == nil {
			return nil
		}
	}
	for an := a.Height; an < b.Height; {
		b, _ := chainIndexerStore.chainStore.GetBlockHeader(&b.PreviousHash)
		if b == nil {
			return nil
		}
	}
	for a.Hash() != b.Hash() {
		a, _ := chainIndexerStore.chainStore.GetBlockHeader(&a.PreviousHash)
		if a == nil {
			return nil
		}
		b, _ := chainIndexerStore.chainStore.GetBlockHeader(&b.PreviousHash)
		if b == nil {
			return nil
		}
	}
	return a
}

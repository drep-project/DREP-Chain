package chain_indexer

import (
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/crypto/sha3"
)

func (chainIndexer *ChainIndexerService) GetStoredSections() uint64 {
	var storedSections uint64
	key := sha3.Keccak256([]byte("bl_storedsections"))
	value, err := chainIndexer.DatabaseService.Get(key)
	if err != nil {
		return storedSections
	}

	err = binary.Unmarshal(value, &storedSections)
	if err != nil {
		return storedSections
	}
	return storedSections
}

func (chainIndexer *ChainIndexerService) SetStoredSections(storedSections uint64) error {
	key := sha3.Keccak256([]byte("bl_storedsections"))
	value, err := binary.Marshal(storedSections)
	if err != nil {
		return err
	}
	return chainIndexer.DatabaseService.Put(key, value)
}
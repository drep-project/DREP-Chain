package chain

import (
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/common/trie"
	"math/big"

	"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/database/memorydb"
	"github.com/drep-project/DREP-Chain/params"
	"github.com/drep-project/DREP-Chain/types"
)

func (chainService *ChainService) GetGenisiBlock(biosAddress crypto.CommonAddress) (*types.Block, error) {
	var root []byte
	db, err := store.TrieStoreFromStore(memorydb.New(), trie.EmptyRoot[:])
	if err != nil {
		return nil, err
	}

	genesisContext, err := NewGenesisContext(&chainService.genesisConfig, db)
	if err != nil {
		return nil, err
	}
	for _, genesisProcess := range chainService.genesisProcess {
		err := genesisProcess.Genesis(genesisContext)
		if err != nil {
			return nil, err
		}
	}

	root = db.GetStateRoot()

	merkleRoot := chainService.DeriveMerkleRoot(nil)
	return &types.Block{
		Header: &types.BlockHeader{
			Version:      common.Version,
			PreviousHash: crypto.Hash{},
			GasLimit:     *new(big.Int).SetUint64(params.GenesisGasLimit),
			GasUsed:      *new(big.Int),
			Timestamp:    1545282765,
			StateRoot:    root,
			TxRoot:       merkleRoot,
			Height:       0,
		},
		Data: &types.BlockData{
			TxCount: 0,
			TxList:  []*types.Transaction{},
		},
	}, nil
}

func (chainService *ChainService) ProcessGenesisBlock(biosAddr crypto.CommonAddress) (*types.Block, error) {
	var err error
	var root []byte

	chainStore, err := store.TrieStoreFromStore(chainService.DatabaseService.LevelDb(), trie.EmptyRoot[:])
	if err != nil {
		return nil, err
	}
	genesisContext, err := NewGenesisContext(&chainService.genesisConfig, chainStore)
	if err != nil {
		return nil, err
	}
	for _, genesisProcess := range chainService.genesisProcess {
		err := genesisProcess.Genesis(genesisContext)
		if err != nil {
			return nil, err
		}
	}

	root = chainStore.GetStateRoot()
	err = chainStore.TrieDB().TrieDb(crypto.Bytes2Hash(root), true)
	if err != nil {
		return nil, err
	}
	merkleRoot := chainService.DeriveMerkleRoot(nil)
	return &types.Block{
		Header: &types.BlockHeader{
			Version:      common.Version,
			PreviousHash: crypto.Hash{},
			GasLimit:     *new(big.Int).SetUint64(params.GenesisGasLimit),
			GasUsed:      *new(big.Int),
			Timestamp:    1545282765,
			StateRoot:    root,
			TxRoot:       merkleRoot,
			Height:       0,
		},
		Data: &types.BlockData{
			TxCount: 0,
			TxList:  []*types.Transaction{},
		},
	}, nil
}

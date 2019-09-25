package chain

import (
	"github.com/drep-project/drep-chain/chain/store"
	"github.com/drep-project/drep-chain/common/trie"
	"math/big"

	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database/memorydb"
	"github.com/drep-project/drep-chain/params"
	"github.com/drep-project/drep-chain/types"
)

func (chainService *ChainService) GetGenisiBlock(biosAddress crypto.CommonAddress) *types.Block {
	var root []byte
	db, _ := store.TrieStoreFromStore(memorydb.New(), trie.EmptyRoot[:])
	for addr, balance := range params.Preminer {
		//add preminer addr and balance
		storage := types.NewStorage()
		storage.Balance = *balance
		db.PutStorage(&addr, storage)
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
	}
}

func (chainService *ChainService) ProcessGenesisBlock(biosAddr crypto.CommonAddress) (*types.Block, error) {
	var err error
	var root []byte

	chainStore, err := store.TrieStoreFromStore(chainService.DatabaseService.LevelDb(), trie.EmptyRoot[:])
	if err != nil {
		return nil, err
	}
	for addr, balance := range params.Preminer {
		//add preminer addr and balance
		storage := types.NewStorage()
		storage.Balance = *balance
		chainStore.PutStorage(&addr, storage)
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

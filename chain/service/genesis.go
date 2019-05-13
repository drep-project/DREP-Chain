package service

import (
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/chain/params"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/database"
	"math/big"
)

func (chainService *ChainService) GetGenisiBlock(biosPubkey string) *chainTypes.Block {
	var root []byte
	db, err := database.DatabaseFromStore(database.NewMemoryStore())
	for addr, balance := range params.Preminer {
		//add preminer addr and balance
		storage := chainTypes.NewStorage()
		storage.Balance = *balance
		db.PutStorage(&addr, storage)
	}
	root = db.GetStateRoot()

	merkleRoot := chainService.deriveMerkleRoot(nil)
	b := common.MustDecode(biosPubkey)
	pubkey, err := secp256k1.ParsePubKey(b)
	if err != nil {
		return nil
	}
	return &chainTypes.Block{
		Header: &chainTypes.BlockHeader{
			Version:      common.Version,
			PreviousHash: crypto.Hash{},
			GasLimit:     *new (big.Int).SetUint64(params.GenesisGasLimit),
			GasUsed:      *new(big.Int),
			Timestamp:    1545282765,
			StateRoot:    root,
			TxRoot:       merkleRoot,
			Height:       0,
			LeaderPubKey: *pubkey,
		},
		Data: &chainTypes.BlockData{
			TxCount: 0,
			TxList:  []*chainTypes.Transaction{},
		},
	}
}

func (chainService *ChainService) ProcessGenesisBlock(genesisPubkey string) (*chainTypes.Block, error) {
	var err error
	var root []byte
	for addr, balance := range params.Preminer {
		//add preminer addr and balance
		storage := chainTypes.NewStorage()
		storage.Balance = *balance
		chainService.DatabaseService.PutStorage(&addr, storage)
	}

	root = chainService.DatabaseService.GetStateRoot()
	if err != nil {
		return nil, err
	}

	merkleRoot := chainService.deriveMerkleRoot(nil)
	b := common.MustDecode(genesisPubkey)
	pubkey, err := secp256k1.ParsePubKey(b)
	if err != nil {
		return nil, err
	}
	chainService.DatabaseService.RecordBlockJournal(0)
	return &chainTypes.Block{
		Header: &chainTypes.BlockHeader{
			Version:      common.Version,
			PreviousHash: crypto.Hash{},
			GasLimit:     *new (big.Int).SetUint64(params.GenesisGasLimit),
			GasUsed:      *new(big.Int),
			Timestamp:    1545282765,
			StateRoot:    root,
			TxRoot:       merkleRoot,
			Height:       0,
			LeaderPubKey: *pubkey,
		},
		Data: &chainTypes.BlockData{
			TxCount: 0,
			TxList:  []*chainTypes.Transaction{},
		},
	}, nil
}
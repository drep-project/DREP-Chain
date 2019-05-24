package chainservice

import (
	chainType "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"math/big"
)

type ChainApi struct {
	chainService *ChainService
	dbService    *database.DatabaseService
}

func (chain *ChainApi) GetBlock(height uint64) (*chainType.RpcBlock, error) {
	blocks, err := chain.chainService.GetBlocksFrom(height, 1)
	if err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, ErrBlockNotFound
	}
	return new(chainType.RpcBlock).From(blocks[0]), nil
}

func (chain *ChainApi) GetMaxHeight() uint64 {
	return chain.chainService.BestChain.Height()
}

func (chain *ChainApi) GetBalance(addr crypto.CommonAddress) *big.Int {
	return chain.dbService.GetBalance(&addr)
}

func (chain *ChainApi) GetNonce(addr crypto.CommonAddress) uint64 {
	return chain.dbService.GetNonce(&addr)
}

func (chain *ChainApi) GetPreviousBlockHash() (string, error) {
	block, err := chain.GetBlock(chain.GetMaxHeight())
	if err != nil {
		return "", err
	}
	return block.PreviousHash.String(), nil
}

func (chain *ChainApi) GetReputation(addr crypto.CommonAddress) *big.Int {
	return chain.dbService.GetReputation(&addr)
}

func (chain *ChainApi) GetTransactionsFromBlock(height uint64) ([]*chainType.RpcTransaction, error) {
	block, err := chain.GetBlock(height)
	if err != nil {
		return nil, err
	}
	return block.Txs, nil
}

func (chain *ChainApi) GetTransactionByBlockHeightAndIndex(height uint64, index int) (*chainType.RpcTransaction, error) {
	block, err := chain.GetBlock(height)
	if err != nil {
		return nil, err
	}
	if index > len(block.Txs) {
		return nil, ErrTxIndexOutOfRange
	}
	return block.Txs[index], nil
}

func (chain *ChainApi) GetTransactionCountByBlockHeight(height uint64) (int, error) {
	block, err := chain.GetBlock(height)
	if err != nil {
		return 0, err
	}
	return len(block.Txs), nil
}



//根据地址获取地址对应的别名
func (chain *ChainApi) GetAliasByAddress(addr *crypto.CommonAddress) string {
	return chain.chainService.DatabaseService.GetStorageAlias(addr)
}

//根据别名获取别名对应的地址
func (chain *ChainApi) GetAddressByAlias(alias string) *crypto.CommonAddress {
	return chain.chainService.DatabaseService.AliasGet(alias)
}
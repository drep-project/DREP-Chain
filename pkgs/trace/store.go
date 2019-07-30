package trace

import (
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/types"
)

// IStore determine the interface to be implementation for storage
type IStore interface {
	ExistRecord(block *types.Block) (bool, error)

	InsertRecord(block *types.Block)

	DelRecord(block *types.Block)

	GetRawTransaction(txHash *crypto.Hash) ([]byte, error)

	GetTransaction(txHash *crypto.Hash) (*types.RpcTransaction, error)

	GetSendTransactionsByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*types.RpcTransaction

	GetReceiveTransactionsByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*types.RpcTransaction

	Close()
}

package trace

import (
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/types"
)

// IStore determine the interface to be implementation for storage
type IStore interface {
	ExistRecord(block *types.Block) (bool, error)

	InsertRecord(block *types.Block)

	DelRecord(block *types.Block)

	GetRawTransaction(txHash *crypto.Hash) ([]byte, error)

	GetTransaction(txHash *crypto.Hash) (*RpcTransaction, error)

	GetSendTransactionsByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*RpcTransaction

	GetReceiveTransactionsByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*RpcTransaction

	Close()
}

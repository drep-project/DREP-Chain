package trace

import (
	"github.com/drep-project/drep-chain/crypto"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
)

// IStore determine the interface to be implementation for storage
type IStore interface {

	InsertRecord(block *chainTypes.Block)

	DelRecord(block *chainTypes.Block)

	GetRawTransaction(txHash *crypto.Hash) ([]byte, error)

	GetTransaction(txHash *crypto.Hash) (*chainTypes.RpcTransaction, error)

	GetSendTransactionsByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*chainTypes.RpcTransaction

    GetReceiveTransactionsByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*chainTypes.RpcTransaction

	Close()
}
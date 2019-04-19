package trace

import (
	"github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/binary"
)

type TraceApi struct {
	service *TraceService
}

func (traceApi *TraceApi) GetRawTrasnaction(txHash *crypto.Hash) (string, error){
	rawData, err := traceApi.service.GetRawTransaction(txHash)
	if err != nil {
		return  "", err
	}
	return common.Encode(rawData), nil
}

func (traceApi *TraceApi) GetTrasnaction(txHash *crypto.Hash) (*types.RpcTransaction, error) {
	tx, err := traceApi.service.GetTransaction(txHash)
	if err != nil {
		return  nil, err
	}
	rpcTx := &types.RpcTransaction{}
	rpcTx.From(tx)
	return rpcTx, nil
}

func (traceApi *TraceApi) DecodeTrasnaction(bytes common.Bytes) (*types.RpcTransaction, error) {
	tx := &types.Transaction{}
	err := binary.Unmarshal(bytes[:], tx)
	if err != nil {
		return  nil, err
	}
	rpcTx := &types.RpcTransaction{}
	rpcTx.From(tx)
	return rpcTx, nil
}

func (traceApi *TraceApi) GetAddrTransaction(addr *crypto.CommonAddress, pageIndex, pageSize int) []*types.RpcTransaction {
	txs := traceApi.service.GetTransactionsByAddr(addr, pageIndex, pageSize)
	rpcTxs := make([]*types.RpcTransaction, len(txs))
	for index, tx := range txs {
		rpcTxs[index] = &types.RpcTransaction{}
		rpcTxs[index].From(tx)
	}
	return rpcTxs
}
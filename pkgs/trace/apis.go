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

func (traceApi *TraceApi) GetRawTransaction(txHash *crypto.Hash) (string, error){
	rawData, err := traceApi.service.store.GetRawTransaction(txHash)
	if err != nil {
		return  "", err
	}
	return common.Encode(rawData), nil
}

func (traceApi *TraceApi) GetTransaction(txHash *crypto.Hash) (*types.RpcTransaction, error) {
	rpcTx, err := traceApi.service.store.GetTransaction(txHash)
	if err != nil {
		return  nil, err
	}
	return rpcTx, nil
}

func (traceApi *TraceApi) DecodeTrasnaction(bytes common.Bytes) (*types.RpcTransaction, error) {
	tx := &types.Transaction{}
	err := binary.Unmarshal(bytes[:], tx)
	if err != nil {
		return  nil, err
	}
	rpcTx := &types.RpcTransaction{}
	rpcTx.FromTx(tx)
	return rpcTx, nil
}

func (traceApi *TraceApi) GetSendTransactionByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*types.RpcTransaction {
	return traceApi.service.store.GetSendTransactionsByAddr(addr, pageIndex, pageSize)
}

func (traceApi *TraceApi) GetReceiveTransactionByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*types.RpcTransaction {
	return traceApi.service.store.GetReceiveTransactionsByAddr(addr, pageIndex, pageSize)
}
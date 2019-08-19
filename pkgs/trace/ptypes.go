package trace

import (
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/types"
	"math/big"
)

type ViewTransaction struct {
	Hash      string
	From      string
	Version   int32
	Nonce     uint64 //交易序列号
	Type      int
	To        string
	ChainId   types.ChainIdType
	Amount    string
	GasPrice  uint64
	GasLimit  uint64
	Timestamp uint64
	Data      string //hex
	Sig       string
	Height    uint64
}

type ViewBlock struct {
	Hash         string
	ChainId      types.ChainIdType
	Version      int32
	PreviousHash string
	GasLimit     uint64
	GasUsed      uint64
	Height       uint64
	Timestamp    uint64
	StateRoot    string //hex
	TxRoot       string //hex
	LeaderPubKey string
	MinorPubKeys []string
	Txs          []string
}
type ViewBlockHeader struct {
	ChainId      types.ChainIdType
	Version      int32
	PreviousHash string
	GasLimit     uint64
	GasUsed      uint64
	Height       uint64
	Timestamp    uint64
	StateRoot    string
	TxRoot       string
	LeaderPubKey string
	MinorPubKeys []string
	Hash         string
}

func (viewBlockHeader *ViewBlockHeader) From(block *types.Block) *ViewBlockHeader {
	txs := make([]*ViewTransaction, len(block.Data.TxList))
	for i, tx := range block.Data.TxList {
		txs[i] = new(ViewTransaction).FromTx(tx)
		txs[i].Height = block.Header.Height
	}

	viewBlockHeader.Hash = block.Header.Hash().String()
	viewBlockHeader.ChainId = block.Header.ChainId
	viewBlockHeader.Version = block.Header.Version
	viewBlockHeader.PreviousHash = block.Header.PreviousHash.String()
	viewBlockHeader.GasLimit = block.Header.GasLimit.Uint64()
	viewBlockHeader.GasUsed = block.Header.GasUsed.Uint64()
	viewBlockHeader.Height = block.Header.Height
	viewBlockHeader.Timestamp = block.Header.Timestamp
	viewBlockHeader.StateRoot = common.Encode(block.Header.StateRoot)
	viewBlockHeader.TxRoot = common.Encode(block.Header.TxRoot)
	//viewBlockHeader.LeaderPubKey = block.Header.LeaderAddress.String()

	viewBlockHeader.MinorPubKeys = []string{}
	//for _, val := range block.Header.MinorAddresses {
	//	viewBlockHeader.MinorPubKeys = append(viewBlockHeader.MinorPubKeys, val.String())
	//}
	return viewBlockHeader
}

func (rpcTransaction *ViewTransaction) FromTx(tx *types.Transaction) *ViewTransaction {
	from, _ := tx.From()
	rpcTransaction.Hash = tx.TxHash().String()
	rpcTransaction.Version = tx.Data.Version
	rpcTransaction.Nonce = tx.Data.Nonce
	rpcTransaction.Type = int(tx.Data.Type)
	rpcTransaction.To = tx.Data.To.String()
	rpcTransaction.ChainId = tx.Data.ChainId
	rpcTransaction.Amount = (*big.Int)(&tx.Data.Amount).String()
	rpcTransaction.GasPrice = (*big.Int)(&tx.Data.GasPrice).Uint64()
	rpcTransaction.GasLimit = (*big.Int)(&tx.Data.GasLimit).Uint64()

	rpcTransaction.Timestamp = uint64(tx.Data.Timestamp)
	if tx.Data.Data != nil {
		rpcTransaction.Data = common.Encode(tx.Data.Data)
	}
	rpcTransaction.From = common.Encode(from.Bytes())
	rpcTransaction.Sig = common.Encode(tx.Sig)
	return rpcTransaction
}

func (rpcBlock *ViewBlock) From(block *types.Block) *ViewBlock {
	txs := make([]*ViewTransaction, len(block.Data.TxList))
	for i, tx := range block.Data.TxList {
		txs[i] = new(ViewTransaction).FromTx(tx)
		txs[i].Height = block.Header.Height
	}

	rpcBlock.Hash = block.Header.Hash().String()
	rpcBlock.ChainId = block.Header.ChainId
	rpcBlock.Version = block.Header.Version
	rpcBlock.PreviousHash = block.Header.PreviousHash.String()
	rpcBlock.GasLimit = block.Header.GasLimit.Uint64()
	rpcBlock.GasUsed = block.Header.GasUsed.Uint64()
	rpcBlock.Height = block.Header.Height
	rpcBlock.Timestamp = uint64(block.Header.Timestamp)
	rpcBlock.StateRoot = common.Encode(block.Header.StateRoot)
	rpcBlock.TxRoot = common.Encode(block.Header.TxRoot)
	//rpcBlock.LeaderPubKey = block.Header.LeaderAddress.String()

	rpcBlock.MinorPubKeys = []string{}
	//for _, val := range block.Header.MinorAddresses {
	//	rpcBlock.MinorPubKeys = append(rpcBlock.MinorPubKeys, val.String())
//	}
	rpcBlock.Txs = make([]string, len(txs))
	for index, val := range txs {
		rpcBlock.Txs[index] = val.Hash
	}
	return rpcBlock
}

type RpcBlockHeader struct {
	ChainId      types.ChainIdType
	Version      int32
	PreviousHash string
	GasLimit     string
	GasUsed      string
	Height       uint64
	Timestamp    uint64
	StateRoot    string
	TxRoot       string
	LeaderPubKey string
	MinorPubKeys []string

	Hash string
}

func (rpcBlockHeader *RpcBlockHeader) FromBlockHeader(header *types.BlockHeader) {
	rpcBlockHeader.ChainId = header.ChainId
	rpcBlockHeader.Version = header.Version
	rpcBlockHeader.PreviousHash = header.PreviousHash.String()
	rpcBlockHeader.GasLimit = header.GasLimit.String()
	rpcBlockHeader.GasUsed = header.GasUsed.String()
	rpcBlockHeader.Height = header.Height
	rpcBlockHeader.Timestamp = header.Timestamp
	rpcBlockHeader.StateRoot = common.Encode(header.StateRoot)
	rpcBlockHeader.TxRoot = common.Encode(header.TxRoot)
	//rpcBlockHeader.LeaderPubKey = header.LeaderAddress.String()
	rpcBlockHeader.MinorPubKeys = []string{}
//	for _, val := range header.MinorAddresses {
//		rpcBlockHeader.MinorPubKeys = append(rpcBlockHeader.MinorPubKeys, val.String())
//	}
	rpcBlockHeader.Hash = header.Hash().String()
}

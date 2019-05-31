package trace

import (
	"github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"strconv"
)

type ViewTransaction struct {
	Hash 	  string
	From 	  string
	Version   string
	Nonce     string //交易序列号
	Type      string
	To        string
	ChainId   string
	Amount    string
	GasPrice  string
	GasLimit  string
	Timestamp int64
	Data      string  //hex
	Sig  	  string
}

type ViewBlock struct {
	Hash         string
	ChainId      string
	Version      string
	PreviousHash string
	GasLimit     string
	GasUsed      string
	Height       string
	Timestamp    string
	StateRoot    string  //hex
	TxRoot       string	 //hex
	LeaderPubKey string
	MinorPubKeys []string
	Txs          []*ViewTransaction
}
type ViewBlockHeader struct {
	ChainId      string
	Version      string
	PreviousHash string
	GasLimit     string
	GasUsed      string
	Height       uint64
	Timestamp    uint64
	StateRoot    string
	TxRoot       string
	LeaderPubKey string
	MinorPubKeys []string
	Hash 	string
}

func (viewBlockHeader *ViewBlockHeader) From(block *types.Block) *ViewBlockHeader {
	txs := make([]*ViewTransaction, len(block.Data.TxList))
	for i, tx := range block.Data.TxList {
		txs[i] = new (ViewTransaction).FromTx(tx)
	}

	viewBlockHeader.Hash 			= block.Header.Hash().String()
	viewBlockHeader.ChainId 		= common.Encode(block.Header.ChainId[:])
	viewBlockHeader.Version 		= strconv.FormatInt(int64(block.Header.Version), 10)
	viewBlockHeader.PreviousHash 	= block.Header.PreviousHash.String()
	viewBlockHeader.GasLimit 		= block.Header.GasLimit.String()
	viewBlockHeader.GasUsed 		= block.Header.GasUsed.String()
	viewBlockHeader.Height 			= block.Header.Height
	viewBlockHeader.Timestamp 		= block.Header.Timestamp
	viewBlockHeader.StateRoot 		= common.Encode(block.Header.StateRoot)
	viewBlockHeader.TxRoot 			= common.Encode(block.Header.TxRoot)
	viewBlockHeader.LeaderPubKey 	= crypto.PubKey2Address(&block.Header.LeaderPubKey).String()

	viewBlockHeader.MinorPubKeys 	= []string{}
	for _, val := range  block.Header.MinorPubKeys {
		viewBlockHeader.MinorPubKeys = append(viewBlockHeader.MinorPubKeys, crypto.PubKey2Address(&val).String())
	}
	return viewBlockHeader
}

func (rpcTransaction *ViewTransaction) FromTx(tx *types.Transaction) *ViewTransaction {
	from, _ := tx.From()
	rpcTransaction.Hash 		= tx.TxHash().String()
	rpcTransaction.Version   	= strconv.FormatInt(int64(tx.Data.Version), 10)
	rpcTransaction.Nonce     	= strconv.FormatUint(uint64(tx.Data.Nonce), 10)
	rpcTransaction.Type      	=  strconv.FormatInt(int64(tx.Data.Type), 10)
	rpcTransaction.To         	= tx.Data.To.String()
	rpcTransaction.ChainId     	= common.Encode(tx.Data.ChainId[:])
	rpcTransaction.Amount      	= tx.Data.Amount.String()
	rpcTransaction.GasPrice    	= tx.Data.GasPrice.String()
	rpcTransaction.GasLimit    	= tx.Data.GasLimit.String()

	rpcTransaction.Timestamp   	= tx.Data.Timestamp
	if 	 tx.Data.Data != nil {
		rpcTransaction.Data = common.Encode(tx.Data.Data)
	}
	rpcTransaction.From = common.Encode(from.Bytes())
	rpcTransaction.Sig =  common.Encode(tx.Sig)
	return rpcTransaction
}

func (rpcBlock *ViewBlock) From(block *types.Block) *ViewBlock {
	txs := make([]*ViewTransaction, len(block.Data.TxList))
	for i, tx := range block.Data.TxList {
		txs[i] = new (ViewTransaction).FromTx(tx)
	}

	rpcBlock.Hash 			= block.Header.Hash().String()
	rpcBlock.ChainId 		= common.Encode(block.Header.ChainId[:])
	rpcBlock.Version 		= strconv.FormatInt(int64( block.Header.Version), 10)
	rpcBlock.PreviousHash 	= block.Header.PreviousHash.String()
	rpcBlock.GasLimit 		= block.Header.GasLimit.String()
	rpcBlock.GasUsed 		= block.Header.GasUsed.String()
	rpcBlock.Height 		= strconv.FormatUint(uint64(block.Header.Height), 10)
	rpcBlock.Timestamp 		= strconv.FormatUint(uint64(block.Header.Timestamp), 10)
	rpcBlock.StateRoot 		= common.Encode(block.Header.StateRoot)
	rpcBlock.TxRoot 		= common.Encode(block.Header.TxRoot)
	rpcBlock.LeaderPubKey 	= crypto.PubKey2Address(&block.Header.LeaderPubKey).String()

	rpcBlock.MinorPubKeys 	= []string{}
	for _, val := range  block.Header.MinorPubKeys {
		rpcBlock.MinorPubKeys = append(rpcBlock.MinorPubKeys, crypto.PubKey2Address(&val).String())
	}
	rpcBlock.Txs = txs
	return rpcBlock
}


type RpcBlockHeader struct {
	ChainId      string
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

	Hash 	string
}


func (rpcBlockHeader *RpcBlockHeader)FromBlockHeader(header *types.BlockHeader){
	rpcBlockHeader.ChainId 		= common.Encode(header.ChainId[:])
	rpcBlockHeader.Version 		= header.Version
	rpcBlockHeader.PreviousHash = header.PreviousHash.String()
	rpcBlockHeader.GasLimit 	= header.GasLimit.String()
	rpcBlockHeader.GasUsed 		= header.GasUsed.String()
	rpcBlockHeader.Height 		= header.Height
	rpcBlockHeader.Timestamp 	= header.Timestamp
	rpcBlockHeader.StateRoot 	= common.Encode(header.StateRoot)
	rpcBlockHeader.TxRoot 		= common.Encode(header.TxRoot)
	rpcBlockHeader.LeaderPubKey =  crypto.PubKey2Address(&header.LeaderPubKey).String()
	rpcBlockHeader.MinorPubKeys =  []string{}
	for _, val := range  header.MinorPubKeys {
		rpcBlockHeader.MinorPubKeys = append(rpcBlockHeader.MinorPubKeys, crypto.PubKey2Address(&val).String())
	}
	rpcBlockHeader.Hash 		= header.Hash().String()
}
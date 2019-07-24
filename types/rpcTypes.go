package types

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"math/big"
)

type RpcTransaction struct {
	Hash            crypto.Hash
	From            crypto.CommonAddress
	TransactionData `bson:",inline"`
	Sig             common.Bytes
}

type RpcBlock struct {
	Hash         crypto.Hash
	ChainId      app.ChainIdType
	Version      int32
	PreviousHash crypto.Hash
	GasLimit     big.Int
	GasUsed      big.Int
	Height       uint64
	Timestamp    uint64
	StateRoot    common.Bytes
	TxRoot       common.Bytes
	LeaderPubKey secp256k1.PublicKey
	MinorPubKeys []secp256k1.PublicKey
	Txs          []*RpcTransaction
}

func (rpcTransaction *RpcTransaction) FromTx(tx *Transaction) *RpcTransaction {
	from, _ := tx.From()
	rpcTransaction.Hash = *tx.TxHash()
	rpcTransaction.TransactionData = tx.Data
	rpcTransaction.From = *from
	rpcTransaction.Sig = common.Bytes(tx.Sig)
	return rpcTransaction
}

func (rpcTx *RpcTransaction) ToTx() *Transaction {
	tx := &Transaction{}
	tx.Data = rpcTx.TransactionData
	tx.Sig = rpcTx.Sig
	return tx
}

func (rpcBlock *RpcBlock) From(block *Block) *RpcBlock {
	txs := make([]*RpcTransaction, len(block.Data.TxList))
	for i, tx := range block.Data.TxList {
		txs[i] = new(RpcTransaction).FromTx(tx)
	}

	rpcBlock.Hash = *block.Header.Hash()
	rpcBlock.ChainId = block.Header.ChainId
	rpcBlock.Version = block.Header.Version
	rpcBlock.PreviousHash = block.Header.PreviousHash
	rpcBlock.GasLimit = block.Header.GasLimit
	rpcBlock.GasUsed = block.Header.GasUsed
	rpcBlock.Height = block.Header.Height
	rpcBlock.Timestamp = block.Header.Timestamp
	rpcBlock.StateRoot = block.Header.StateRoot
	rpcBlock.TxRoot = block.Header.TxRoot
	rpcBlock.LeaderPubKey = block.Header.LeaderPubKey
	rpcBlock.MinorPubKeys = block.Header.MinorPubKeys
	rpcBlock.Txs = txs
	return rpcBlock
}

type RpcBlockHeader struct {
	ChainId      app.ChainIdType
	Version      int32
	PreviousHash crypto.Hash
	GasLimit     common.Big
	GasUsed      common.Big
	Height       uint64
	Timestamp    uint64
	StateRoot    []byte
	TxRoot       []byte
	LeaderPubKey secp256k1.PublicKey
	MinorPubKeys []secp256k1.PublicKey

	Hash *crypto.Hash
}

func (rpcBlockHeader *RpcBlockHeader) FromBlockHeader(header *BlockHeader) {
	rpcBlockHeader.ChainId = header.ChainId
	rpcBlockHeader.Version = header.Version
	rpcBlockHeader.PreviousHash = header.PreviousHash
	rpcBlockHeader.GasLimit = common.Big(header.GasLimit)
	rpcBlockHeader.GasUsed = common.Big(header.GasUsed)
	rpcBlockHeader.Height = header.Height
	rpcBlockHeader.Timestamp = header.Timestamp
	rpcBlockHeader.StateRoot = header.StateRoot
	rpcBlockHeader.TxRoot = header.TxRoot
	rpcBlockHeader.LeaderPubKey = header.LeaderPubKey
	rpcBlockHeader.MinorPubKeys = header.MinorPubKeys
	rpcBlockHeader.Hash = header.Hash()
}

func (rpcBlockHeader *RpcBlockHeader) ToHeader() *BlockHeader {
	blockHeader := &BlockHeader{}
	blockHeader.ChainId = rpcBlockHeader.ChainId
	blockHeader.Version = rpcBlockHeader.Version
	blockHeader.PreviousHash = rpcBlockHeader.PreviousHash
	blockHeader.GasLimit = (big.Int)(rpcBlockHeader.GasLimit)
	blockHeader.GasUsed = (big.Int)(rpcBlockHeader.GasUsed)
	blockHeader.Height = rpcBlockHeader.Height
	blockHeader.Timestamp = rpcBlockHeader.Timestamp
	blockHeader.StateRoot = rpcBlockHeader.StateRoot
	blockHeader.TxRoot = rpcBlockHeader.TxRoot
	blockHeader.LeaderPubKey = rpcBlockHeader.LeaderPubKey
	blockHeader.MinorPubKeys = rpcBlockHeader.MinorPubKeys
	return blockHeader
}

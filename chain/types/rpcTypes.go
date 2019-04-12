package types

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"math/big"
)

type RpcTransaction struct {
	Hash crypto.Hash
	Data TransactionData
	Sig  common.Bytes
}

type RpcBlock struct {
	Hash        crypto.Hash
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
	Txs         []*RpcTransaction
}

func (rpcTransaction *RpcTransaction) From(tx *Transaction) *RpcTransaction {
	rpcTransaction.Hash = *tx.TxHash()
	rpcTransaction.Data = tx.Data
	rpcTransaction.Sig = common.Bytes(tx.Sig)
	return rpcTransaction
}

func (rpcBlock *RpcBlock) From(block *Block) *RpcBlock {
	txs := make([]*RpcTransaction, len(block.Data.TxList))
	for i, tx := range block.Data.TxList {
		txs[i] = new (RpcTransaction).From(tx)
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
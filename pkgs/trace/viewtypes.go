package trace

import (
	"encoding/hex"
	"fmt"
	"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/pkgs/consensus/service/bft"
	consensusTypes "github.com/drep-project/DREP-Chain/pkgs/consensus/types"
	"github.com/drep-project/DREP-Chain/types"
	"github.com/drep-project/binary"
	"math/big"
)

type ViewTransaction struct {
	Id        string `bson:"_id"`
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
	Id           string `bson:"_id"`
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
	Proof        interface{}
	Txs          []string
}

type ViewBlockHeader struct {
	Id           string `bson:"_id"`
	ChainId      types.ChainIdType
	Version      int32
	PreviousHash string
	GasLimit     uint64
	GasUsed      uint64
	Height       uint64
	Timestamp    uint64
	StateRoot    string
	TxRoot       string
	Hash         string
}

func (viewBlockHeader *ViewBlockHeader) From(block *types.Block) *ViewBlockHeader {
	txs := make([]*ViewTransaction, len(block.Data.TxList))
	for i, tx := range block.Data.TxList {
		txs[i] = new(ViewTransaction).FromTx(tx)
		txs[i].Height = block.Header.Height
	}

	viewBlockHeader.Id = block.Header.Hash().String()
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
	return viewBlockHeader
}

func (rpcTransaction *ViewTransaction) FromTx(tx *types.Transaction) *ViewTransaction {
	from, _ := tx.From()
	rpcTransaction.Id = tx.TxHash().String()
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

func (rpcBlock *ViewBlock) From(block *types.Block, addresses []crypto.CommonAddress) *ViewBlock {
	txs := make([]*ViewTransaction, len(block.Data.TxList))
	for i, tx := range block.Data.TxList {
		txs[i] = new(ViewTransaction).FromTx(tx)
		txs[i].Height = block.Header.Height
	}
	rpcBlock.Id = block.Header.Hash().String()
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

	if block.Proof.Type == consensusTypes.Solo {
		rpcBlock.Proof = block.Proof
	} else if block.Proof.Type == consensusTypes.Pbft {
		proof := NewPbftProof()
		multiSig := &bft.MultiSignature{}
		binary.Unmarshal(block.Proof.Evidence, multiSig)
		proof.Evidence = hex.EncodeToString(block.Proof.Evidence)
		if len(addresses) <= multiSig.Leader {
			fmt.Printf("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&:%s,%d", addresses, multiSig.Leader)
			return nil
		}
		proof.LeaderAddress = addresses[multiSig.Leader].String()
		for index, val := range multiSig.Bitmap {
			if val == 1 {
				proof.MinorAddresses = append(proof.MinorAddresses, addresses[index].String())
			}
		}
		rpcBlock.Proof = proof
	}

	rpcBlock.Txs = make([]string, len(txs))
	for index, val := range txs {
		rpcBlock.Txs[index] = val.Hash
	}
	return rpcBlock
}

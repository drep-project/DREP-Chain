package trace

import (
	"encoding/hex"
	"encoding/json"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/pkgs/consensus/service/bft"
	types2 "github.com/drep-project/drep-chain/pkgs/consensus/types"
	"github.com/drep-project/drep-chain/types"
	"math/big"
)

type RpcTransaction struct {
	Hash                  crypto.Hash
	From                  crypto.CommonAddress
	types.TransactionData `bson:",inline"`
	Sig                   common.Bytes
}

type RpcBlock struct {
	Hash         crypto.Hash
	ChainId      types.ChainIdType
	Version      int32
	PreviousHash crypto.Hash
	GasLimit     big.Int
	GasUsed      big.Int
	Height       uint64
	Timestamp    uint64
	StateRoot    common.Bytes
	TxRoot       common.Bytes
	Txs          []*RpcTransaction
	Proof        interface{}
}

func (rpcTransaction *RpcTransaction) FromTx(tx *types.Transaction) *RpcTransaction {
	from, _ := tx.From()
	rpcTransaction.Hash = *tx.TxHash()
	rpcTransaction.TransactionData = tx.Data
	rpcTransaction.From = *from
	rpcTransaction.Sig = common.Bytes(tx.Sig)
	return rpcTransaction
}

func (rpcTx *RpcTransaction) ToTx() *types.Transaction {
	tx := &types.Transaction{}
	tx.Data = rpcTx.TransactionData
	tx.Sig = rpcTx.Sig
	return tx
}

func (rpcBlock *RpcBlock) From(block *types.Block, addresses []crypto.CommonAddress) *RpcBlock {
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
	rpcBlock.Txs = txs

	if block.Proof.Type == types2.Solo {
		rpcBlock.Proof = block.Proof
	} else if block.Proof.Type == types2.Pbft {
		proof := NewPbftProof()
		multiSig := &bft.MultiSignature{}
		json.Unmarshal(block.Proof.Evidence, multiSig)
		proof.Evidence = hex.EncodeToString(block.Proof.Evidence)
		proof.LeaderPubKey = addresses[multiSig.Leader]
		for index, val := range multiSig.Bitmap {
			if val == 1 {
				proof.MinorPubKeys = append(proof.MinorPubKeys, addresses[index])
			}
		}
		rpcBlock.Proof = proof
	}
	return rpcBlock
}

type PbftProof struct {
	Type         int
	LeaderPubKey crypto.CommonAddress
	MinorPubKeys []crypto.CommonAddress
	Evidence     string
}

func NewPbftProof() *PbftProof {
	return &PbftProof{
		Type:         types2.Pbft,
		MinorPubKeys: []crypto.CommonAddress{},
	}
}

type SoloProof struct {
	Type     int
	Evidence string
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
	gasLimut := common.Big(header.GasLimit)
	rpcBlockHeader.GasLimit = (&gasLimut).String()
	gasUsed := common.Big(header.GasUsed)
	rpcBlockHeader.GasUsed = (&gasUsed).String()
	rpcBlockHeader.Height = header.Height
	rpcBlockHeader.Timestamp = header.Timestamp
	rpcBlockHeader.StateRoot = hex.EncodeToString(header.StateRoot)
	rpcBlockHeader.TxRoot = hex.EncodeToString(header.TxRoot)
	rpcBlockHeader.Hash = header.Hash().String()
}

func (rpcBlockHeader *RpcBlockHeader) ToHeader() *types.BlockHeader {
	blockHeader := &types.BlockHeader{}
	blockHeader.ChainId = rpcBlockHeader.ChainId
	blockHeader.Version = rpcBlockHeader.Version
	blockHeader.PreviousHash = crypto.HexToHash(rpcBlockHeader.PreviousHash)
	blockHeader.GasLimit = *common.MustDecodeBig(rpcBlockHeader.GasLimit)
	blockHeader.GasUsed = *common.MustDecodeBig(rpcBlockHeader.GasUsed)
	blockHeader.Height = rpcBlockHeader.Height
	blockHeader.Timestamp = rpcBlockHeader.Timestamp
	blockHeader.StateRoot = mustDecode(rpcBlockHeader.StateRoot)
	blockHeader.TxRoot = mustDecode(rpcBlockHeader.TxRoot)
	return blockHeader
}

func mustDecode(str string) []byte {
	strBytes, _ := hex.DecodeString(str)
	return strBytes
}

package types

import (
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"math/big"
)

type BlockHeader struct {
	ChainId        ChainIdType
	Version        int32
	PreviousHash   crypto.Hash
	GasLimit       big.Int
	GasUsed        big.Int
	Height         uint64
	Timestamp      uint64
	StateRoot      []byte
	TxRoot         []byte
	ReceiptRoot  crypto.Hash
	Bloom		 Bloom
	LeaderAddress  crypto.CommonAddress
	MinorAddresses []crypto.CommonAddress

	blockHash *crypto.Hash `binary:"ignore"`
}

func (blockHeader *BlockHeader) Hash() *crypto.Hash {
	if blockHeader.blockHash == nil {
		b, err := binary.Marshal(blockHeader)
		if err != nil {
			return nil
		}
		bytes := sha3.Keccak256(b)
		blockHeader.blockHash = &crypto.Hash{}
		blockHeader.blockHash.SetBytes(bytes)
	}
	return blockHeader.blockHash
}

type BlockData struct {
	TxCount uint64
	TxList  []*Transaction
}

type Block struct {
	Header   *BlockHeader
	Data     *BlockData
	MultiSig *MultiSignature
}

func (block *Block) GasUsed() uint64 {
	return block.Header.GasUsed.Uint64()
}

func (block *Block) GasLimit() uint64 {
	return block.Header.GasLimit.Uint64()
}

func (block *Block) AsSignMessage() []byte {
	blockTemp := &Block{
		Header: &BlockHeader{
			ChainId:       block.Header.ChainId,
			Version:       block.Header.Version,
			PreviousHash:  block.Header.PreviousHash,
			GasLimit:      block.Header.GasLimit,
			GasUsed:       block.Header.GasUsed,
			Height:        block.Header.Height,
			Timestamp:     block.Header.Timestamp,
			TxRoot:        block.Header.TxRoot,
			LeaderAddress: block.Header.LeaderAddress,
		},
	}
	bytes, _ := binary.Marshal(blockTemp)
	return bytes
}

func (block *Block) AsMessage() []byte {
	bytes, _ := binary.Marshal(block)
	return bytes
}

func BlockFromMessage(bytes []byte) (*Block, error) {
	block := &Block{}
	err := binary.Unmarshal(bytes, block)
	if err != nil {
		return nil, err
	}
	return block, nil
}

type MultiSignature struct {
	Sig    secp256k1.Signature
	Bitmap []byte
}

func (multiSignature *MultiSignature) AsSignMessage() []byte {
	bytes, _ := binary.Marshal(multiSignature)
	return bytes
}

func (multiSignature *MultiSignature) AsMessage() []byte {
	return multiSignature.AsSignMessage()
}

func MultiSignatureFromMessage(bytes []byte) (*MultiSignature, error) {
	multySig := &MultiSignature{}
	err := binary.Unmarshal(bytes, multySig)
	if err != nil {
		return nil, err
	}
	return multySig, nil
}

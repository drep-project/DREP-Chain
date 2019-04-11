package types

import (
	"encoding/hex"
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"math/big"
)

type BlockHeader struct {
	ChainId      app.ChainIdType        `json:"chainid"    gencodec:"required"`
	Version      uint64                 `json:"version"    gencodec:"required"`
	PreviousHash *crypto.Hash           `json:"prehash"    gencodec:"required"`
	GasLimit     *big.Int               `json:"gaslimit"   gencodec:"required"`
	GasUsed      *big.Int               `json:"gasused"    gencodec:"required"`
	Height       uint64                 `json:"height"     gencodec:"required"`
	Timestamp    uint64                 `json:"timestamp"  gencodec:"required"`
	StateRoot    []byte                 `json:"stateroot"  gencodec:"required"`
	TxRoot       []byte                 `json:"txroot"     gencodec:"required"`
	LeaderPubKey *secp256k1.PublicKey   `json:"leaderpubkey"  gencodec:"required"` //存储pubkey序列化后的值
	MinorPubKeys []*secp256k1.PublicKey `json:"minorpubkey"   gencodec:"required"`
}

func (blockHeader *BlockHeader) Hash() *crypto.Hash {
	b, err := binary.Marshal(blockHeader)
	if err != nil {
		return nil
	}
	bytes := sha3.Hash256(b)
	hash := crypto.Hash{}
	hash.SetBytes(bytes)
	return &hash
}

func (blockHeader *BlockHeader) String() string {
	h := blockHeader.Hash()
	return "0x" + hex.EncodeToString(h[:])
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

func (block *Block) ToMessage() []byte {
	blockTemp := &Block{
		Header: &BlockHeader{
			ChainId:      block.Header.ChainId,
			Version:      block.Header.Version,
			PreviousHash: block.Header.PreviousHash,
			GasLimit:     block.Header.GasLimit,
			GasUsed:      block.Header.GasUsed,
			Height:       block.Header.Height,
			Timestamp:    block.Header.Timestamp,
			StateRoot:    block.Header.StateRoot,
			TxRoot:       block.Header.TxRoot,
			LeaderPubKey: block.Header.LeaderPubKey,
		},
		Data: block.Data,
	}
	bytes, _ := binary.Marshal(blockTemp)
	return bytes
}

type MultiSignature struct {
	Sig    secp256k1.Signature
	Bitmap []byte
}

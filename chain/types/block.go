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
	ChainId      app.ChainIdType
	Version      int32
	PreviousHash *crypto.Hash
	GasLimit     *big.Int
	GasUsed      *big.Int
	Height       int64
	Timestamp    int64
	StateRoot    []byte
	TxRoot       []byte
	LeaderPubKey *secp256k1.PublicKey
	MinorPubKeys []*secp256k1.PublicKey
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
	TxCount int32
	TxList  []*Transaction
}

type Block struct {
	Header   *BlockHeader
	Data     *BlockData
	MultiSig *MultiSignature
}

func (block *Block) ToMessage() []byte {
	blockTemp := &Block{
		Header:&BlockHeader{
			ChainId  	 :block.Header.ChainId,
			Version      :block.Header.Version,
			PreviousHash :block.Header.PreviousHash,
			GasLimit     :block.Header.GasLimit,
			GasUsed      :block.Header.GasUsed,
			Height       :block.Header.Height,
			Timestamp    :block.Header.Timestamp,
			StateRoot    :block.Header.StateRoot,
			TxRoot       :block.Header.TxRoot,
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

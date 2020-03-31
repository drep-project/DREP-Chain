package types

import (
	"fmt"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"github.com/drep-project/binary"
	"math/big"
)

type BlockHeader struct {
	ChainId      ChainIdType
	Version      int32
	PreviousHash crypto.Hash
	GasLimit     big.Int
	GasUsed      big.Int
	Height       uint64
	Timestamp    uint64
	StateRoot    []byte
	TxRoot       []byte
	ReceiptRoot  crypto.Hash
	Bloom        Bloom
	blockHash    *crypto.Hash `binary:"ignore"`
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
	Header *BlockHeader
	Data   *BlockData
	Proof  Proof
}

type Proof struct {
	Type     int
	Evidence []byte
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
			ChainId:      block.Header.ChainId,
			Version:      block.Header.Version,
			PreviousHash: block.Header.PreviousHash,
			GasLimit:     block.Header.GasLimit,
			GasUsed:      block.Header.GasUsed,
			Height:       block.Header.Height,
			Timestamp:    block.Header.Timestamp,
			TxRoot:       block.Header.TxRoot,
			ReceiptRoot:  block.Header.ReceiptRoot,
			Bloom:        block.Header.Bloom,
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
	if len(bytes) > 10*1024*1024 {
		return nil, fmt.Errorf("bytes to len:%d", len(bytes))
	}
	err := binary.Unmarshal(bytes, block)
	if err != nil {
		fmt.Println("BlockFromMessage err:", err)
		return nil, err
	}

	if block.Header == nil {
		fmt.Println("header len :", len(bytes))
		fmt.Println("fatal err,block parse err **************************************")
		//panic("header is nil")
		return nil, fmt.Errorf("unmarshal err")
	}

	return block, nil
}

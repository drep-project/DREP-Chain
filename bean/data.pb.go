package bean

import "BlockChainTest/mycrypto"

type BlockHeader struct {
	Version              int32
	PreviousHash         []byte
	GasLimit             []byte
	GasUsed              []byte
	Height               int64
	Timestamp            int64
	MerkleRoot           []byte
	TxHashes             [][]byte
	LeaderPubKey         *mycrypto.Point
	MinorPubKeys         []*mycrypto.Point
}

type BlockData struct {
	TxCount              int32
	TxList               []*Transaction
}

type Block struct {
	Header               *BlockHeader
	Data                 *BlockData
	MultiSig             *mycrypto.Signature
}

type TransactionData struct {
	Version              int32
	Nonce                int64
	Type                 int32
	To                   string
	Amount               []byte
	GasPrice             []byte
	GasLimit             []byte
	Timestamp            int64
	Data                 []byte
	PubKey               *mycrypto.Point
}

type Transaction struct {
	Data                 *TransactionData
	Sig                  *mycrypto.Signature
}

type MultiSignature struct {
	Sig                  *mycrypto.Signature
	Bitmap               []byte
}
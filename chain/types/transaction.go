package types

import (
	"encoding/json"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"math/big"
	"time"
	"github.com/drep-project/binary"
)

type Transaction struct {
	Data TransactionData
	Sig  []byte
}

type TransactionData struct {
	Version   int32
	Nonce     uint64 //交易序列号
	Type      TxType
	To        crypto.CommonAddress
	ChainId   app.ChainIdType
	Amount    big.Int
	GasPrice  big.Int
	GasLimit  big.Int
	Timestamp int64
	Data      []byte
	From      crypto.CommonAddress
}

func (tx *Transaction) Nonce() uint64 {
	return tx.Data.Nonce
}

func (tx *Transaction) Type() TxType {
	return tx.Data.Type
}
func (tx *Transaction) Gas() uint64        { return tx.Data.GasLimit.Uint64() }

func (tx *Transaction) From() *crypto.CommonAddress {
	return &tx.Data.From
}

type CrossChainTransaction struct {
	ChainId   app.ChainIdType
	StateRoot []byte
	Trans     []*Transaction
}

func (tx *Transaction) GetData() []byte {
	return tx.Data.Data
}

func (tx *Transaction) Cost() *big.Int {
	total := new(big.Int).Mul(&tx.Data.GasPrice, new(big.Int).SetUint64(tx.Data.GasLimit.Uint64()))
	total.Add(total, &tx.Data.Amount)
	return total
}

func (tx *Transaction) To() *crypto.CommonAddress {
	return &tx.Data.To
}

func (tx *Transaction) ChainId() app.ChainIdType {
	return tx.Data.ChainId
}

func (tx *Transaction) Amount() *big.Int {
	return &tx.Data.Amount
}

func (tx *Transaction) GasLimit() *big.Int {
	return &tx.Data.GasLimit
}
func (tx *Transaction) GasPrice() *big.Int {
	return &tx.Data.GasPrice
}

//func (tx *Transaction) PubKey() *secp256k1.PublicKey {
//	return tx.Data.PubKey
//}

func (tx *Transaction) TxHash() *crypto.Hash {
	b, _ := binary.Marshal(tx.Data)
	h := sha3.Hash256(b)
	hash := crypto.Hash{}
	hash.SetBytes(h)
	return &hash
}

func (tx *Transaction) TxSig(prvKey *secp256k1.PrivateKey) (*secp256k1.Signature, error) {
	b, err := binary.Marshal(tx.Data)
	if err != nil {
		return nil, err
	}

	return prvKey.Sign(sha3.Hash256(b))
}

func (tx *Transaction) GetGasUsed() *big.Int {
	return new(big.Int).SetInt64(int64(100))
}

func (tx *Transaction) GetGas() *big.Int {
	gasQuantity := tx.GetGasUsed()
	gasUsed := new(big.Int).Mul(gasQuantity, &tx.Data.GasPrice)
	return gasUsed
}

func (tx *Transaction) GetSig() []byte {
	return tx.Sig
}

func (tx *Transaction) AsPersistentMessage() []byte {
	txBytes, _ := json.Marshal(tx)
	return txBytes
}

type Message struct {
	To         *crypto.CommonAddress
	From       *crypto.CommonAddress
	Nonce      uint64
	Amount     *big.Int
	GasLimit   uint64
	GasPrice   *big.Int
	Data       []byte
	CheckNonce bool
}

func NewTransaction(from crypto.CommonAddress, to crypto.CommonAddress, amount *big.Int, nonce uint64) *Transaction {
	data := TransactionData{
		Version:   common.Version,
		Nonce:     nonce,
		Type:      TransferType,
		To:        to,
		Amount:    *amount,
		GasPrice:  *DefaultGasPrice,
		GasLimit:  *TransferGas,
		Timestamp: time.Now().Unix(),
		From:      from,
	}
	return &Transaction{Data: data}
}

func NewContractTransaction(from crypto.CommonAddress, byteCode []byte, nonce uint64) *Transaction {
	data := TransactionData{
		Nonce:     nonce,
		Type:      CreateContractType,
		GasPrice:  *DefaultGasPrice,
		GasLimit:  *CreateContractGas,
		Timestamp: time.Now().Unix(),
		Data:      byteCode,
		From:      from,
	}
	return &Transaction{Data: data}
}

func NewCallContractTransaction(from crypto.CommonAddress, to crypto.CommonAddress, input []byte, amount *big.Int, nonce uint64) *Transaction {
	data := TransactionData{
		Nonce:     nonce,
		Type:      CallContractType,
		To:        to,
		Amount:    *amount,
		GasPrice:  *DefaultGasPrice,
		GasLimit:  *CallContractGas,
		Timestamp: time.Now().Unix(),
		From:      from,
		Data:      input,
	}
	return &Transaction{Data: data}
}

package types

import (
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"math/big"
	"sync/atomic"
	"time"
)

type Transaction struct {
	Data TransactionData
	Sig  []byte

	txHash      *crypto.Hash `json:"-" binary:"ignore"`
	signMessage []byte       `json:"-" binary:"ignore" bson:"-"`
	message     []byte       `json:"-" binary:"ignore" bson:"-"`
	from        atomic.Value `json:"-" binary:"ignore"`
}

type TransactionData struct {
	Version   int32
	Nonce     uint64 //交易序列号
	Type      TxType
	To        crypto.CommonAddress
	ChainId   app.ChainIdType
	Amount    common.Big
	GasPrice  common.Big
	GasLimit  common.Big
	Timestamp int64
	Data      []byte
}

func (tx *Transaction) Nonce() uint64 {
	return tx.Data.Nonce
}

func (tx *Transaction) Type() TxType {
	return tx.Data.Type
}
func (tx *Transaction) Gas() uint64 {
	bigInt := (big.Int)(tx.Data.GasLimit)
	return (&bigInt).Uint64()
}

func (tx *Transaction) From() (*crypto.CommonAddress, error) {
	if sc := tx.from.Load(); sc != nil {
		return sc.(*crypto.CommonAddress), nil
	}

	pk, _, err := secp256k1.RecoverCompact(tx.Sig, tx.TxHash().Bytes())
	if err != nil {
		return nil, err
	}
	addr := crypto.PubKey2Address(pk)
	tx.from.Store(&addr)
	return &addr, nil
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
	total := new(big.Int).Mul(tx.GasPrice(), new(big.Int).SetUint64(tx.Gas()))
	total.Add(total, tx.Amount())
	return total
}

func (tx *Transaction) To() *crypto.CommonAddress {
	return &tx.Data.To
}

func (tx *Transaction) ChainId() app.ChainIdType {
	return tx.Data.ChainId
}

func (tx *Transaction) Amount() *big.Int {
	bigint := (big.Int)(tx.Data.Amount)
	return &bigint
}

func (tx *Transaction) GasLimit() *big.Int {
	bigint := (big.Int)(tx.Data.GasLimit)
	return &bigint
}
func (tx *Transaction) GasPrice() *big.Int {
	bigint := (big.Int)(tx.Data.GasPrice)
	return &bigint
}

//func (tx *Transaction) PubKey() *secp256k1.PublicKey {
//	return tx.Data.PubKey
//}

func (tx *Transaction) TxHash() *crypto.Hash {
	if tx.txHash == nil {
		b := tx.AsSignMessage()
		h := sha3.Hash256(b)
		tx.txHash = &crypto.Hash{}
		tx.txHash.SetBytes(h)
	}
	return tx.txHash
}

func (tx *Transaction) GetSig() []byte {
	return tx.Sig
}

func (tx *Transaction) AsSignMessage() []byte {
	if tx.signMessage == nil {
		tx.signMessage, _ = binary.Marshal(tx.Data)
	}
	return tx.signMessage
}

func (tx *Transaction) AsPersistentMessage() []byte {
	if tx.message == nil {
		tx.message, _ = binary.Marshal(tx)
	}
	return tx.message
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

func NewTransaction(to crypto.CommonAddress, amount, gasPrice, gasLimit *big.Int, nonce uint64) *Transaction {
	data := TransactionData{
		Version:   common.Version,
		Nonce:     nonce,
		Type:      TransferType,
		To:        to,
		Amount:    *(*common.Big)(amount),
		GasPrice:  *(*common.Big)(gasPrice),
		GasLimit:  *(*common.Big)(gasLimit),
		Timestamp: time.Now().Unix(),
	}
	return &Transaction{Data: data}
}

func NewContractTransaction(byteCode []byte, gasPrice, gasLimit *big.Int, nonce uint64, ) *Transaction {
	if gasPrice == nil {
		gasPrice = DefaultGasPrice
	}
	if gasLimit == nil {
		gasLimit = CreateContractGas
	}
	data := TransactionData{
		Nonce:     nonce,
		Type:      CreateContractType,
		GasPrice:  *(*common.Big)(gasPrice),
		GasLimit:  *(*common.Big)(gasLimit),
		Timestamp: time.Now().Unix(),
		Data:      byteCode,
	}
	return &Transaction{Data: data}
}

func NewCallContractTransaction(to crypto.CommonAddress, input []byte, amount, gasPrice, gasLimit *big.Int, nonce uint64) *Transaction {
	if gasPrice == nil {
		gasPrice = DefaultGasPrice
	}
	if gasLimit == nil {
		gasLimit = CallContractGas
	}
	data := TransactionData{
		Nonce:     nonce,
		Type:      CallContractType,
		To:        to,
		Amount:    *(*common.Big)(amount),
		GasPrice:  *(*common.Big)(gasPrice),
		GasLimit:  *(*common.Big)(gasLimit),
		Timestamp: time.Now().Unix(),
		Data:      input,
	}
	return &Transaction{Data: data}
}

//给地址srcAddr设置别名
func NewAliasTransaction(alias string, gasPrice, gasLimit *big.Int, nonce uint64) *Transaction {
	data := TransactionData{
		Version:   common.Version,
		Nonce:     nonce,
		Type:      SetAliasType,
		To:        crypto.CommonAddress{},
		Amount:    *(*common.Big)(new(big.Int)),
		GasPrice:  *(*common.Big)(gasPrice),
		GasLimit:  *(*common.Big)(gasLimit),
		Timestamp: int64(time.Now().Unix()),
		Data:      []byte(alias),
	}
	return &Transaction{Data: data}
}

package types

import (
"encoding/hex"
"encoding/json"
"github.com/drep-project/drep-chain/app"
"github.com/drep-project/drep-chain/common"
"github.com/drep-project/drep-chain/crypto"
"github.com/drep-project/drep-chain/crypto/secp256k1"
"github.com/drep-project/drep-chain/crypto/sha3"
"math/big"
"time"
)

type Transaction struct {
	data *TransactionData
	Sig  []byte
}

type TransactionData struct {
	Version   int32
	Nonce     int64 //交易序列号
	Type      TxType
	To        crypto.CommonAddress
	ChainId   app.ChainIdType
	Amount    *big.Int
	GasPrice  *big.Int
	GasLimit  *big.Int
	Timestamp int64
	Data      []byte
	From      crypto.CommonAddress
}

func (tx *Transaction) Nonce() int64 {
	return tx.data.Nonce
}

func (tx *Transaction) Type() TxType {
	return tx.data.Type
}

func (tx *Transaction) From() *crypto.CommonAddress {
	return &tx.data.From
}

type CrossChainTransaction struct {
	ChainId   app.ChainIdType
	StateRoot []byte
	Trans     []*Transaction
}

func (tx *Transaction) Data() []byte {
	return tx.data.Data
}

func (tx *Transaction) To() *crypto.CommonAddress {
	return &tx.data.To
}

func (tx *Transaction) ChainId() app.ChainIdType {
	return tx.data.ChainId
}

func (tx *Transaction) Amount() *big.Int {
	return tx.data.Amount
}
func (tx *Transaction) GasLimit() *big.Int {
	return tx.data.GasLimit
}
func (tx *Transaction) GasPrice() *big.Int {
	return tx.data.GasPrice
}

//func (tx *Transaction) PubKey() *secp256k1.PublicKey {
//	return tx.data.PubKey
//}

func (tx *Transaction) TxId() (string, error) {
	b, err := json.Marshal(tx.data)
	if err != nil {
		return "", err
	}
	id := hex.EncodeToString(sha3.Hash256(b))
	return id, nil
}

func (tx *Transaction) TxHash() ([]byte, error) {
	b, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}
	h := sha3.Hash256(b)
	return h, nil
}

func (tx *Transaction) TxSig(prvKey *secp256k1.PrivateKey) (*secp256k1.Signature, error) {
	b, err := json.Marshal(tx.data)
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
	gasUsed := new(big.Int).Mul(gasQuantity, tx.data.GasPrice)
	return gasUsed
}

func NewTransaction(from crypto.CommonAddress, to crypto.CommonAddress, amount *big.Int, nonce int64) *Transaction {
	data := &TransactionData{
		Version:   common.Version,
		Nonce:     nonce,
		Type:      TransferType,
		To:        to,
		Amount:    amount,
		GasPrice:  DefaultGasPrice,
		GasLimit:  TransferGas,
		Timestamp: time.Now().Unix(),
		From:      from,
	}
	return &Transaction{data: data}
}

func NewContractTransaction(from crypto.CommonAddress, to crypto.CommonAddress, byteCode []byte, nonce int64) *Transaction {
	nonce++
	data := &TransactionData{
		Nonce:     nonce,
		Type:      CreateContractType,
		GasPrice:  DefaultGasPrice,
		GasLimit:  CreateContractGas,
		Timestamp: time.Now().Unix(),
		Data:      make([]byte, len(byteCode)+1),
		From:      from,
	}
	copy(data.Data[1:], byteCode)
	data.Data[0] = 2
	return &Transaction{data: data}
}

func NewCallContractTransaction(from crypto.CommonAddress, to crypto.CommonAddress, input []byte, amount *big.Int, nonce int64, readOnly bool) *Transaction {
	nonce++
	data := &TransactionData{
		Nonce:     nonce,
		Type:      CallContractType,
		To:        to,
		Amount:    amount,
		GasPrice:  DefaultGasPrice,
		GasLimit:  CallContractGas,
		Timestamp: time.Now().Unix(),
		From:      from,
		Data:      make([]byte, len(input)+1),
	}
	copy(data.Data[1:], input)
	if readOnly {
		data.Data[0] = 1
	} else {
		data.Data[0] = 0
	}
	return &Transaction{data: data}
}

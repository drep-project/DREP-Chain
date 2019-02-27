package types

import (
	"encoding/hex"
	"encoding/json"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"math/big"
)

type Transaction struct {
	Data                 *TransactionData
	Sig                  []byte
}


type TransactionData struct {
	Version              int32
	Nonce                int64
	Type                 int32
	To                   crypto.CommonAddress
	ChainId              app.ChainIdType
	Amount               *big.Int
	GasPrice             *big.Int
	GasLimit             *big.Int
	Timestamp            int64
	Data                 []byte
	PubKey               *secp256k1.PublicKey
}

type CrossChainTransaction struct {
	ChainId   app.ChainIdType
	StateRoot []byte
	Trans     []*Transaction
}

func (tx *Transaction) TxId() (string, error) {
	b, err := json.Marshal(tx.Data)
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
	b, err := json.Marshal(tx.Data)
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
	gasUsed := new(big.Int).Mul(gasQuantity, tx.Data.GasPrice)
	return gasUsed
}
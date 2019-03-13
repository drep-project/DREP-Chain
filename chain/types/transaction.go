package types

import (
	"encoding/hex"
	"encoding/json"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"math/big"
)

type Transaction struct {
	Data *TransactionData
	Sig  []byte
}

type TransactionData struct {
	Version   int32
	Nonce     int64 //交易序列号
	Type      TxType
	GasPrice  *big.Int
	GasLimit  *big.Int
	ChainId   app.ChainIdType
	Timestamp int64
	FromAccount      string
	Comment   string
	Amount    *big.Int
	Data      []byte
}

type TransferAction struct {
	To        string
}

// UnmarshalJSON parses a hash in hex syntax.
func (transferAction TransferAction) UnmarshalJSON(input []byte) error {
	transferAction.To = string(input)
	return nil
}

func (transferAction TransferAction) MarshalText() ([]byte, error) {
	return []byte(transferAction.To), nil
}


func NewTransferAction(toAccount string) *TransferAction{
	return &TransferAction{
		To: 	toAccount,
	}
}

type RegisterMinerAction struct {
	MinerAccount string
	SignKey		 secp256k1.PublicKey

}

// UnmarshalJSON parses a hash in hex syntax.
func (registerMinerAction RegisterMinerAction) UnmarshalJSON(input []byte) error {
	return nil
}

func (registerMinerAction RegisterMinerAction) MarshalText() ([]byte, error) {
	return []byte(""), nil
}

func NewRegisterMinerAction(minerAccount string, signKey secp256k1.PublicKey) *RegisterMinerAction{
	return &RegisterMinerAction{
		MinerAccount: 	minerAccount,
		SignKey: signKey,
	}
}


type RegisterAccountAction struct {
	Name   		string
	Authority	Authority
	ChainId    	app.ChainIdType
	ChainCode  	[]byte
}

// UnmarshalJSON parses a hash in hex syntax.
func (registerAccountAction RegisterAccountAction) UnmarshalJSON(input []byte) error {
	return nil
}

func (registerAccountAction RegisterAccountAction) MarshalText() ([]byte, error) {
	return []byte(""), nil
}

func NewRegisterAccountAction(name string, authority Authority, chainId app.ChainIdType, chainCode []byte) *RegisterAccountAction{
	return &RegisterAccountAction{
		Name: 		name,
		Authority: 	authority,
		ChainId: 	chainId,
		ChainCode: 	chainCode,
	}
}

type CallContractAction struct {
	ContractName    string
	Input			[]byte
	Readonly		bool
}

func (callContractAction CallContractAction) UnmarshalJSON(input []byte) error {
	return nil
}

func (callContractAction CallContractAction) MarshalText() ([]byte, error) {
	return []byte(""), nil
}

func NewCallContractAction(contractName string, input []byte, readOnly bool) *CallContractAction {
	return &CallContractAction{
		ContractName: 	contractName,
		Input: 			input,
		Readonly:		readOnly,
	}
}

/*
func NewCallContractTransaction(from, to string, input []byte, amount *big.Int, nonce int64, readOnly bool) *Transaction {
	nonce++
	data := &TransactionData{
		Nonce:     nonce,
		Type:      CallContractType,
		From:      from,
		GasPrice:  DefaultGasPrice,
		GasLimit:  CallContractGas,
		Timestamp: time.Now().Unix(),
		Data:      make([]byte, len(input)+1),
		To:        to,
		Amount:    amount,
	}
	copy(data.Data[1:], input)
	if readOnly {
		data.Data[0] = 1
	} else {
		data.Data[0] = 0
	}
	return &Transaction{Data: data}
}
*/


type CreateContractAction struct {
	ContractName    string
	ByteCode		[]byte
}

func (createContractAction CreateContractAction) UnmarshalJSON(input []byte) error {
	return nil
}

func (createContractAction CreateContractAction) MarshalText() ([]byte, error) {
	return []byte(""), nil
}

func NewCallCreateContractAction(contractName string, byteCode []byte) *CreateContractAction {
	return &CreateContractAction{
		ContractName: 	contractName,
		ByteCode: 		byteCode,
	}
}

type CrossChainAction struct {
	ChainId   app.ChainIdType
	StateRoot []byte
	Trans     []*Transaction
}

/*
func NewCreateContractTransaction(from, to string, byteCode []byte, nonce int64) *Transaction {
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
	return &Transaction{Data: data}
}
*/

func (tx *Transaction) Nonce() int64 {
	return tx.Data.Nonce
}

func (tx *Transaction) Type() TxType {
	return tx.Data.Type
}

func (tx *Transaction) Amount() *big.Int {
	return tx.Data.Amount
}


func (tx *Transaction) From() string {
	return tx.Data.FromAccount
}

func (tx *Transaction) GetData() []byte {
	return tx.Data.Data
}

func (tx *Transaction) ChainId() app.ChainIdType {
	return tx.Data.ChainId
}

func (tx *Transaction) GasLimit() *big.Int {
	return tx.Data.GasLimit
}
func (tx *Transaction) GasPrice() *big.Int {
	return tx.Data.GasPrice
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
	b, err := json.Marshal(tx.Data)
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

func (tx *Transaction) GetSig() []byte {
	return tx.Sig
}

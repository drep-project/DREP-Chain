package types

import (
	"errors"
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"math/big"
	"time"
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

func NewTransaction(from string, txType TxType ,amount *big.Int, nonce int64, gasPrice, gasLimit *big.Int, action interface{}) (*Transaction, error) {
	actionBytes, err := binary.Marshal(action)
	if err != nil {
		return nil, err
	}
	data := &TransactionData{
		Version:   		common.Version,
		FromAccount:    from,
		Type:      		txType,
		Amount:    		amount,
		Nonce:     		nonce,
		GasPrice:  		gasPrice,
		GasLimit:  		gasLimit,
		Data:			actionBytes,
		Timestamp: 		time.Now().Unix(),
	}
	return &Transaction{Data: data}, nil
}


type TransferAction struct {
	To        string
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

func NewCreateContractAction(contractName string, byteCode []byte) *CreateContractAction {
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

func (tx *Transaction) TxHash() crypto.Hash {
	b, _ := binary.Marshal(tx.Data)
	h := crypto.Hash{}
	h.SetBytes(sha3.Hash256(b))
	return h
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
	gasUsed := new(big.Int).Mul(gasQuantity, tx.Data.GasPrice)
	return gasUsed
}

func (tx *Transaction) GetSig() []byte {
	return tx.Sig
}

type Message struct {
	Type     TxType
	From      string
	ChainId   app.ChainIdType
	DestChain app.ChainIdType
	Gas       *big.Int
	Value     *big.Int
	Nonce     uint64
	Action		interface{}
}

func TxToMessage(tx *Transaction) (*Message, error) {
	var action interface{}

	switch tx.Type() {
	case CreateContractType:
		action = &CreateContractAction{}
		err := binary.Unmarshal(tx.GetData(), action)
		if err != nil{
			return  nil, err
		}
	case CallContractType:
		action = &CallContractAction{}
		err := binary.Unmarshal(tx.GetData(), action)
		if err != nil{
			return  nil, err
		}
	default:
		return  nil, errors.New("unsupport type")
	}

	return &Message{
		From:      tx.From(),
		ChainId:   tx.ChainId(),
		Gas:       tx.GasLimit(),
		Value:     tx.Amount(),
		Nonce:     uint64(tx.Nonce()),
	}, nil
}
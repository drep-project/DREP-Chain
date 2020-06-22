package types

import (
	"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"github.com/drep-project/DREP-Chain/params"
	"github.com/drep-project/binary"
	"math"
	"math/big"
	"sync/atomic"
	"time"
)

type Transaction struct {
	Data TransactionData
	Sig  []byte

	txHash      atomic.Value `json:"-" binary:"ignore"`
	signMessage atomic.Value `json:"-" binary:"ignore" bson:"-"`
	message     atomic.Value `json:"-" binary:"ignore" bson:"-"`
	from        atomic.Value `json:"-" binary:"ignore"`
}

type TransactionData struct {
	Version   int32
	Nonce     uint64 //seq of transaction
	Type      TxType
	To        crypto.CommonAddress
	ChainId   ChainIdType
	Amount    common.Big
	GasPrice  common.Big
	GasLimit  common.Big
	Timestamp int64
	Data      []byte
}

func (tx *Transaction) Time() int64 {
	return tx.Data.Timestamp
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
	addr := crypto.PubkeyToAddress(pk)
	tx.from.Store(&addr)
	return &addr, nil
}

type CrossChainTransaction struct {
	ChainId   ChainIdType
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

func (tx *Transaction) ChainId() ChainIdType {
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

func (tx *Transaction) TxHash() *crypto.Hash {
	if val := tx.txHash.Load(); val != nil {
		return val.(*crypto.Hash)
	}

	b := tx.AsSignMessage()
	h := sha3.Keccak256(b)
	txHash := &crypto.Hash{}
	txHash.SetBytes(h)
	tx.txHash.Store(txHash)
	return txHash
}

func (tx *Transaction) GetSig() []byte {
	return tx.Sig
}

func (tx *Transaction) AsSignMessage() []byte {
	if val := tx.signMessage.Load(); val != nil {
		return val.([]byte)
	}
	signMessage, _ := binary.Marshal(tx.Data)
	tx.signMessage.Store(signMessage)
	return signMessage
}

func (tx *Transaction) AsPersistentMessage() []byte {
	if val := tx.message.Load(); val != nil {
		return val.([]byte)
	}
	message, _ := binary.Marshal(tx)
	tx.message.Store(message)
	return message
}

func (tx *Transaction) IntrinsicGas() (uint64, error) {
	data := tx.AsPersistentMessage()
	contractCreation := (tx.To() == nil || tx.To().IsEmpty()) && tx.Type() != SetAliasType
	// Set the starting gas for the raw transaction
	var gas uint64
	if contractCreation {
		gas = params.TxGasContractCreation
	} else {
		gas = params.TxGas
	}
	// Bump the required gas by the amount of transactional data
	if len(data) > 0 {
		// Zero and non-zero bytes are priced differently
		var nz uint64
		for _, byt := range data {
			if byt != 0 {
				nz++
			}
		}
		// Make sure we don't exceed uint64 for all data combinations
		if (math.MaxUint64-gas)/params.TxDataNonZeroGas < nz {
			return 0, ErrOutOfGas
		}
		gas += nz * params.TxDataNonZeroGas

		z := uint64(len(data)) - nz
		if (math.MaxUint64-gas)/params.TxDataZeroGas < z {
			return 0, ErrOutOfGas
		}
		gas += z * params.TxDataZeroGas
	}
	return gas, nil
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

func NewContractTransaction(byteCode []byte, gasPrice, gasLimit *big.Int, nonce uint64) *Transaction {
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

//Set an alias to the address srcAddr
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

func NewVoteTransaction(to crypto.CommonAddress, amount, gasPrice, gasLimit *big.Int, nonce uint64) *Transaction {
	data := TransactionData{
		Version:   common.Version,
		Nonce:     nonce,
		Type:      VoteCreditType,
		To:        to,
		Amount:    *(*common.Big)(amount),
		GasPrice:  *(*common.Big)(gasPrice),
		GasLimit:  *(*common.Big)(gasLimit),
		Timestamp: int64(time.Now().Unix()),
	}
	return &Transaction{Data: data}
}

func NewCancelVoteTransaction(to crypto.CommonAddress, amount, gasPrice, gasLimit *big.Int, nonce uint64) *Transaction {
	data := TransactionData{
		Version:   common.Version,
		Nonce:     nonce,
		Type:      CancelVoteCreditType,
		To:        to,
		Amount:    *(*common.Big)(amount),
		GasPrice:  *(*common.Big)(gasPrice),
		GasLimit:  *(*common.Big)(gasLimit),
		Timestamp: int64(time.Now().Unix()),
	}
	return &Transaction{Data: data}
}

func NewCandidateTransaction(amount, gasPrice, gasLimit *big.Int, nonce uint64, data []byte) *Transaction {
	txData := TransactionData{
		Version:   common.Version,
		Nonce:     nonce,
		Type:      CandidateType,
		Amount:    *(*common.Big)(amount),
		GasPrice:  *(*common.Big)(gasPrice),
		GasLimit:  *(*common.Big)(gasLimit),
		Timestamp: int64(time.Now().Unix()),
		Data:      data,
	}
	return &Transaction{Data: txData}
}

func NewCancleCandidateTransaction(amount, gasPrice, gasLimit *big.Int, nonce uint64) *Transaction {
	txData := TransactionData{
		Version:   common.Version,
		Nonce:     nonce,
		Type:      CancelCandidateType,
		Amount:    *(*common.Big)(amount),
		GasPrice:  *(*common.Big)(gasPrice),
		GasLimit:  *(*common.Big)(gasLimit),
		Timestamp: int64(time.Now().Unix()),
	}
	return &Transaction{Data: txData}
}

type ExecuteTransactionResult struct {
	TxResult              []byte               //Transaction execution results
	ContractTxExecuteFail bool                 //contract transaction execution results
	ContractTxLog         []*Log               //contract transaction execution logs
	Txerror               error                //transaction execution fail info
	ContractAddr          crypto.CommonAddress //create new contract address
}

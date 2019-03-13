package vm

import (
	"errors"
	"github.com/drep-project/drep-chain/app"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"math/big"
	"sync"
)

var (
	state                   *State
	once                    sync.Once
	ErrNotAccountAddress    = errors.New("a non account address occupied")
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrAccountNotExists     = errors.New("account not exists")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrCodeAlreadyExists    = errors.New("code already exists")
	ErrCodeNotExists        = errors.New("code not exists")
	ErrNotLogAddress        = errors.New("a non log address occupied")
	ErrLogAlreadyExists     = errors.New("log already exists")
)

type State struct {
	databaseApi *database.DatabaseService
	refund      uint64
}

func NewState(databaseService *database.DatabaseService) *State {
	return &State{}
}


func (s *State) CreateContractAccount(callerName string, chainId app.ChainIdType, nonce int64) (*chainTypes.Account, error) {
	account, err := chainTypes.NewContractAccount(callerName, chainId)
	if err != nil {
		return nil, err
	}
	return account, s.databaseApi.PutStorage(account.Name, account.Storage, true)
}


func (s *State) SubBalance(accountName string, amount *big.Int) error {
	balance := s.databaseApi.GetBalance(accountName, true)
	return s.databaseApi.PutBalance(accountName, new(big.Int).Sub(balance, amount), true)
}

func (s *State) AddBalance(accountName string, amount *big.Int) error {
	s.databaseApi.AddBalance(accountName, amount, true)
	//balance := s.databaseApi.GetBalance(addr, chainId,true)
	//return s.databaseApi.PutBalance(addr, chainId, new(big.Int).Add(balance, amount),true)
	return nil
}

func (s *State) GetBalance(accountName string,) *big.Int {
	return s.databaseApi.GetBalance(accountName, true)
}

func (s *State) SetNonce(accountName string, nonce int64) error {
	return s.databaseApi.PutNonce(accountName, nonce, true)
}

func (s *State) GetNonce(accountName string,) int64 {
	return s.databaseApi.GetNonce(accountName, true)
}

func (s *State) Suicide(accountName string,) error {
	storage := s.databaseApi.GetStorage(accountName, true)
	storage.Balance = new(big.Int)
	storage.Nonce = 0
	return s.databaseApi.PutStorage(accountName, storage, true)
}

func (s *State) GetByteCode(accountName string,) crypto.ByteCode {
	return s.databaseApi.GetByteCode(accountName, true)
}

func (s *State) GetCodeSize(accountName string,) int {
	byteCode := s.GetByteCode(accountName)
	return len(byteCode)

}

func (s *State) GetCodeHash(accountName string,) crypto.Hash {
	return s.databaseApi.GetCodeHash(accountName, true)
}

func (s *State) SetByteCode(accountName string, byteCode crypto.ByteCode) error {
	return s.databaseApi.PutByteCode(accountName, byteCode, true)
}

func (s *State) GetLogs(txHash []byte,) []*chainTypes.Log {
	return s.databaseApi.GetLogs(txHash)
}

func (s *State) AddLog(contractName string, txHash, data []byte, topics [][]byte) error {
	log := &chainTypes.Log{
		Name: contractName,
		TxHash:  txHash,
		Data:    data,
		Topics:  topics,
	}
	return s.databaseApi.AddLog(log)
}

func (s *State) AddRefund(gas uint64) {
	s.refund += gas
}

func (s *State) SubRefund(gas uint64) {
	if gas > s.refund {
		panic("refund below zero")
	}
	s.refund -= gas
}

func (s *State) Load(x *big.Int) []byte {
	return s.databaseApi.Load(x)
}

func (s *State) Store(x, y *big.Int) {
	s.databaseApi.Store(x, y)
}

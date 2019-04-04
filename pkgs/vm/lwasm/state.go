package lwasm

import (
	"errors"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"math/big"
)

var (
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

func (s *State) SubBalance(accountName string, amount *big.Int) error {
	balance, err := s.databaseApi.GetBalance(accountName, true)
	if err != nil {
		return err
	}
	return s.databaseApi.PutBalance(accountName, new(big.Int).Sub(balance, amount), true)
}

func (s *State) AddBalance(accountName string, amount *big.Int) error {
	s.databaseApi.AddBalance(accountName, amount, true)
	//balance := s.databaseApi.GetBalance(addr, chainId,true)
	//return s.databaseApi.PutBalance(addr, chainId, new(big.Int).Add(balance, amount),true)
	return nil
}

func (s *State) GetBalance(accountName string,) (*big.Int, error) {
	return s.databaseApi.GetBalance(accountName, true)
}

func (s *State) GetReputation(accountName string,) *big.Int {
	return s.databaseApi.GetReputation(accountName, true)
}

func (s *State) SetNonce(accountName string, nonce int64) error {
	return s.databaseApi.PutNonce(accountName, nonce, true)
}

func (s *State) GetNonce(accountName string,) int64 {
	return s.databaseApi.GetNonce(accountName, true)
}

func (s *State) Suicide(accountName string,) error {
	storage, _ := s.databaseApi.GetStorage(accountName, true)
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

func (s *State) GetLogs(txHash []byte,) []chainTypes.Log {
	return s.databaseApi.GetLogs(txHash)
}

func (s *State) AddLog(contractName string, txHash, data []byte, topics [][]byte) error {
	log := chainTypes.Log{
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

func (s *State) Load(x []byte) []byte {
	return s.databaseApi.Load(x)
}

func (s *State) Store(x, y []byte) {
	s.databaseApi.Store(x, y)
}

func (s *State) Delete(x []byte) {
	s.databaseApi.Delete(x)
}

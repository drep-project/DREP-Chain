package vm

import (
	"BlockChainTest/database"
	"sync"
	"math/big"
	"errors"
	"fmt"
	"BlockChainTest/accounts"
	"BlockChainTest/bean"
)

var (
	state *State
	once sync.Once
	ErrNotAccountAddress = errors.New("a non account address occupied")
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrAccountNotExists = errors.New("account not exists")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrCodeAlreadyExists = errors.New("code already exists")
	ErrCodeNotExists = errors.New("code not exists")
	ErrNotLogAddress = errors.New("a non log address occupied")
	ErrLogAlreadyExists = errors.New("log already exists")
)

type State struct {
	db *database.Database
	refund uint64
}

func NewState() *State {
	return &State{
		db:   database.GetDatabase(),
	}
}

func GetState() *State {
	once.Do(func() {
		if state == nil {
			state = NewState()
		}
	})
	return state
}

func (s *State) CreateContractAccount(callerAddr accounts.CommonAddress, chainId, nonce int64) (*accounts.Account, error) {
	account, err := accounts.NewContractAccount(callerAddr, chainId, nonce)
	if err != nil {
		return nil, err
	}
	return account, database.PutStorage(account.Address, chainId, account.Storage)
}

func (s *State) SubBalance(addr accounts.CommonAddress, chainId int64, amount *big.Int) error {
	balance := database.GetBalance(addr, chainId)
	return database.PutBalance(addr, chainId, new(big.Int).Sub(balance, amount))
}

func (s *State) AddBalance(addr accounts.CommonAddress, chainId int64, amount *big.Int) error {
	balance := database.GetBalance(addr, chainId)
	return database.PutBalance(addr, chainId, new(big.Int).Add(balance, amount))
}

func (s *State) GetBalance(addr accounts.CommonAddress, chainId int64) *big.Int {
	return database.GetBalance(addr, chainId)
}

func (s *State) SetNonce(addr accounts.CommonAddress, chainId int64, nonce int64) error {
	return database.PutNonce(addr, chainId, nonce)
}

func (s *State) GetNonce(addr accounts.CommonAddress, chainId int64) int64 {
	return database.GetNonce(addr, chainId)
}

func (s *State) Suicide(addr accounts.CommonAddress, chainId int64) error {
	storage := database.GetStorage(addr, chainId)
	storage.Balance = new(big.Int)
	storage.Nonce = 0
	return database.PutStorage(addr, chainId, storage)
}

func (s *State) GetByteCode(addr accounts.CommonAddress, chainId int64) accounts.ByteCode {
	storage := database.GetStorage(addr, chainId)
	return storage.ByteCode
}

func (s *State) GetCodeSize(addr accounts.CommonAddress, chainId int64) int {
	byteCode := s.GetByteCode(addr, chainId)
	return len(byteCode)
}

func (s *State) GetCodeHash(addr accounts.CommonAddress, chainId int64) []byte {
	storage := database.GetStorage(addr, chainId)
	return storage.CodeHash.Bytes()
}

func (s *State) SetByteCode(addr accounts.CommonAddress, chainId int64, byteCode accounts.ByteCode) error {
	storage := database.GetStorage(addr, chainId)
	storage.ByteCode = byteCode
	storage.CodeHash = accounts.GetByteCodeHash(byteCode)
	return database.PutStorage(addr, chainId, storage)
}

func (s *State) GetLogs(txHash []byte) []*bean.Log {
	return database.GetLogs(txHash)
}

func (s *State) AddLog(contractAddr accounts.CommonAddress, chainId int64, txHash, data []byte, topics [][]byte) error {
	log := &bean.Log{
		Address: contractAddr,
		ChainId: chainId,
		TxHash: txHash,
		Data: data,
		Topics: topics,
	}
	return database.AddLog(log)
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
	v, err := s.db.Load(x.Bytes())
	if err != nil {
		return new(big.Int).Bytes()
	}
	return v
}

func (s *State) Store(x, y *big.Int) error {
	fmt.Println("x: ", x, " y: ", y)
	return s.db.Store(x.Bytes(), y.Bytes())
}

func (s *State) GetDB() *database.Database {
	return s.db
}
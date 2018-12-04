package vm

import (
	"BlockChainTest/database"
	"sync"
	"math/big"
	"errors"
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
	transaction *database.Transaction
	refund      uint64
}

func NewState() *State {
	return &State{}
}

func GetState() *State {
	state := NewState()
	state.transaction = database.BeginTransaction()
	return state
}

func (s *State) CreateContractAccount(callerAddr accounts.CommonAddress, chainId, nonce int64) (*accounts.Account, error) {
	account, err := accounts.NewContractAccount(callerAddr, chainId, nonce)
	if err != nil {
		return nil, err
	}
	return account, database.PutStorageInsideTransaction(s.transaction, account.Storage, account.Address, chainId)
}

func (s *State) SubBalance(addr accounts.CommonAddress, chainId int64, amount *big.Int) error {
	balance := database.GetBalanceInsideTransaction(s.transaction, addr, chainId)
	return database.PutBalanceInsideTransaction(s.transaction, addr, chainId, new(big.Int).Sub(balance, amount))
}

func (s *State) AddBalance(addr accounts.CommonAddress, chainId int64, amount *big.Int) error {
	balance := database.GetBalanceInsideTransaction(s.transaction, addr, chainId)
	return database.PutBalanceInsideTransaction(s.transaction, addr, chainId, new(big.Int).Add(balance, amount))
}

func (s *State) GetBalance(addr accounts.CommonAddress, chainId int64) *big.Int {
	return database.GetBalanceInsideTransaction(s.transaction, addr, chainId)
}

func (s *State) SetNonce(addr accounts.CommonAddress, chainId int64, nonce int64) error {
	return database.PutNonceInsideTransaction(s.transaction, addr, chainId, nonce)
}

func (s *State) GetNonce(addr accounts.CommonAddress, chainId int64) int64 {
	return database.GetNonceInsideTransaction(s.transaction, addr, chainId)
}

func (s *State) Suicide(addr accounts.CommonAddress, chainId int64) error {
	storage := database.GetStorageInsideTransaction(s.transaction, addr, chainId)
	storage.Balance = new(big.Int)
	storage.Nonce = 0
	return database.PutStorageInsideTransaction(s.transaction, storage, addr, chainId)
}

func (s *State) GetByteCode(addr accounts.CommonAddress, chainId int64) accounts.ByteCode {
	return database.GetByteCodeInsideTransaction(s.transaction, addr, chainId)
}

func (s *State) GetCodeSize(addr accounts.CommonAddress, chainId int64) int {
	byteCode := s.GetByteCode(addr, chainId)
	return len(byteCode)
}

func (s *State) GetCodeHash(addr accounts.CommonAddress, chainId int64) accounts.Hash {
	return database.GetCodeHashInsideTransaction(s.transaction, addr, chainId)
}

func (s *State) SetByteCode(addr accounts.CommonAddress, chainId int64, byteCode accounts.ByteCode) error {
	return database.PutByteCodeInsideTransaction(s.transaction, addr, chainId, byteCode)
}

func (s *State) GetLogs(txHash []byte, chainId int64) []*bean.Log {
	return database.GetLogsInsideTransaction(s.transaction, txHash, chainId)
}

func (s *State) AddLog(contractAddr accounts.CommonAddress, chainId int64, txHash, data []byte, topics [][]byte) error {
	log := &bean.Log{
		Address: contractAddr,
		ChainId: chainId,
		TxHash: txHash,
		Data: data,
		Topics: topics,
	}
	return database.AddLogInsideTransaction(s.transaction, log)
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
	value := s.transaction.Get(x.Bytes())
	if value == nil {
		return new(big.Int).Bytes()
	}
	return value
}

func (s *State) Store(x, y *big.Int, chainId int64) {
	s.transaction.Put(x.Bytes(), y.Bytes(), chainId)
}

func (s *State) Commit() {
	s.transaction.Commit()
}
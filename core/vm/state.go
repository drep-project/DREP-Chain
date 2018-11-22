package vm

import (
	"BlockChainTest/database"
	"sync"
	"math/big"
	"errors"
	"fmt"
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

func (s *State) CreateContractAccount(addr bean.CommonAddress, byteCode bean.ByteCode) error {
	codeHash := bean.CodeHash(byteCode)
	account := &bean.Account{
		Addr: addr.Bytes(),
		Balance: new(big.Int),
		Nonce:   0,
		ByteCode: byteCode,
		CodeHash: codeHash.Bytes(),
	}
	return database.PutAccount(account)
}

func (s *State) SubBalance(addr bean.CommonAddress, amount *big.Int) error {
	balance := database.GetBalance(addr)
	return database.PutBalance(addr, new(big.Int).Sub(balance, amount))
}

func (s *State) AddBalance(addr bean.CommonAddress, amount *big.Int) error {
	balance := database.GetBalance(addr)
	return database.PutBalance(addr, new(big.Int).Add(balance, amount))
}

func (s *State) GetBalance(addr bean.CommonAddress) *big.Int {
	return database.GetBalance(addr)
}

func (s *State) SetNonce(addr bean.CommonAddress, nonce int64) error {
	return database.PutNonce(addr, nonce)
}

func (s *State) GetNonce(addr bean.CommonAddress) int64 {
	return database.GetNonce(addr)
}


func (s *State) Suicide(addr bean.CommonAddress) error {
	account := database.GetAccount(addr)
	account.Nonce = 0
	account.Balance = new(big.Int)
	return database.PutAccount(account)
}

func (s *State) GetAccountStorage(x *big.Int) bean.CommonAddress {
	addr := bean.Big2Address(x)
	err := database.GetAccount(addr)
	if err != nil {
		return bean.CommonAddress{}
	}
	return addr
}


func (s *State) GetByteCode(addr bean.CommonAddress) bean.ByteCode {
	account := database.GetAccount(addr)
	return account.ByteCode
}

func (s *State) GetCodeSize(addr bean.CommonAddress) int {
	account := database.GetAccount(addr)
	return len(account.ByteCode)
}

func (s *State) GetCodeHash(addr bean.CommonAddress) []byte {
	account := database.GetAccount(addr)
	return account.CodeHash
}

func (s *State) GetLog(contractAddr bean.CommonAddress, txHash []byte) *bean.Log {
	log := &bean.Log{ContractAddr: contractAddr.Bytes(), TxHash: txHash}
	addr := log.Address()
	return database.GetLog(addr)
}

func (s *State) AddLog(callerAddr, contractAddr bean.CommonAddress, txHash, data []byte, topics [][]byte) error {
	log := &bean.Log{
		CallerAddr: callerAddr.Bytes(),
		ContractAddr: contractAddr.Bytes(),
		TxHash: txHash,
		Data: data,
		Topics: topics,
	}
	return database.PutLog(log)
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
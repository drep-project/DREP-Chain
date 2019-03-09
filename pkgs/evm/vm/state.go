package vm

import (
	"errors"
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


func (s *State) CreateContractAccount(callerAddr crypto.CommonAddress, nonce int64) (*chainTypes.Account, error) {
	account, err := chainTypes.NewContractAccount(callerAddr, nonce)
	if err != nil {
		return nil, err
	}
	return account, s.databaseApi.CacheStorage(account.Address, account.Storage)
}


func (s *State) SubBalance(addr *crypto.CommonAddress, amount *big.Int) error {
	balance := s.databaseApi.GetBalance(addr)
	return s.databaseApi.CacheBalance(addr, new(big.Int).Sub(balance, amount))
}

func (s *State) AddBalance(addr *crypto.CommonAddress, amount *big.Int) error {
	s.databaseApi.AddBalance(addr, amount)
	//balance := s.databaseApi.GetBalance(addr, chainId,true)
	//return s.databaseApi.PutBalance(addr, chainId, new(big.Int).Add(balance, amount),true)
	return nil
}

func (s *State) GetBalance(addr *crypto.CommonAddress,) *big.Int {
	return s.databaseApi.GetBalance(addr)
}

func (s *State) SetNonce(addr *crypto.CommonAddress, nonce int64) error {
	return s.databaseApi.CacheNonce(addr, nonce)
}

func (s *State) GetNonce(addr *crypto.CommonAddress,) int64 {
	return s.databaseApi.GetNonce(addr)
}

func (s *State) Suicide(addr *crypto.CommonAddress,) error {
	storage := s.databaseApi.GetStorage(addr)
	storage.Balance = new(big.Int)
	storage.Nonce = 0
	return s.databaseApi.CacheStorage(addr, storage)
}

func (s *State) GetByteCode(addr *crypto.CommonAddress,) crypto.ByteCode {
	return s.databaseApi.GetCachedByteCode(addr)
}

func (s *State) GetCodeSize(addr crypto.CommonAddress,) int {
	byteCode := s.GetByteCode(&addr)
	return len(byteCode)

}

func (s *State) GetCodeHash(addr crypto.CommonAddress,) crypto.Hash {
	return s.databaseApi.GetCachedCodeHash(&addr)
}

func (s *State) SetByteCode(addr *crypto.CommonAddress, byteCode crypto.ByteCode) error {
	return s.databaseApi.CacheByteCode(addr, byteCode)
}

func (s *State) GetLogs(txHash []byte,) []*chainTypes.Log {
	return s.databaseApi.GetLogs(txHash)
}

func (s *State) AddLog(contractAddr crypto.CommonAddress, txHash, data []byte, topics [][]byte) error {
	log := &chainTypes.Log{
		Address: contractAddr,
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

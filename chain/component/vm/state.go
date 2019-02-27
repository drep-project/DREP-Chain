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

func (s *State) CreateContractAccount(callerAddr crypto.CommonAddress, chainId app.ChainIdType, nonce int64) (*chainTypes.Account, error) {
	account, err := chainTypes.NewContractAccount(callerAddr, chainId, nonce)
	if err != nil {
		return nil, err
	}
	return account, s.databaseApi.PutStorage(*account.Address, account.Storage, true)
}

func (s *State) SubBalance(addr crypto.CommonAddress, chainId app.ChainIdType, amount *big.Int) error {
	balance := s.databaseApi.GetBalance(addr, chainId, true)
	return s.databaseApi.PutBalance(addr, chainId, new(big.Int).Sub(balance, amount), true)
}

func (s *State) AddBalance(addr crypto.CommonAddress, chainId app.ChainIdType, amount *big.Int) error {
	s.databaseApi.AddBalance(addr, amount, chainId, true)
	//balance := s.databaseApi.GetBalance(addr, chainId,true)
	//return s.databaseApi.PutBalance(addr, chainId, new(big.Int).Add(balance, amount),true)
	return nil
}

func (s *State) GetBalance(addr crypto.CommonAddress, chainId app.ChainIdType) *big.Int {
	return s.databaseApi.GetBalance(addr, chainId, true)
}

func (s *State) SetNonce(addr crypto.CommonAddress, chainId app.ChainIdType, nonce int64) error {
	return s.databaseApi.PutNonce(addr, nonce, true)
}

func (s *State) GetNonce(addr crypto.CommonAddress, chainId app.ChainIdType) int64 {
	return s.databaseApi.GetNonce(addr, true)
}

func (s *State) Suicide(addr crypto.CommonAddress, chainId app.ChainIdType) error {
	storage := s.databaseApi.GetStorage(addr, true)
	storage.Balance = new(big.Int)
	storage.Nonce = 0
	return s.databaseApi.PutStorage(addr, storage, true)
}

func (s *State) GetByteCode(addr crypto.CommonAddress, chainId app.ChainIdType) crypto.ByteCode {
	return s.databaseApi.GetByteCode(addr, true)
}

func (s *State) GetCodeSize(addr crypto.CommonAddress, chainId app.ChainIdType) int {
	byteCode := s.GetByteCode(addr, chainId)
	return len(byteCode)

}

func (s *State) GetCodeHash(addr crypto.CommonAddress, chainId app.ChainIdType) crypto.Hash {
	return s.databaseApi.GetCodeHash(addr, chainId, true)
}

func (s *State) SetByteCode(addr crypto.CommonAddress, chainId app.ChainIdType, byteCode crypto.ByteCode) error {
	return s.databaseApi.PutByteCode(addr, byteCode, true)
}

func (s *State) GetLogs(txHash []byte, chainId app.ChainIdType) []*chainTypes.Log {
	return s.databaseApi.GetLogs(txHash, chainId)
}

func (s *State) AddLog(contractAddr crypto.CommonAddress, chainId app.ChainIdType, txHash, data []byte, topics [][]byte) error {
	log := &chainTypes.Log{
		Address: contractAddr,
		ChainId: chainId,
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

func (s *State) Store(x, y *big.Int, chainId app.ChainIdType) {
	s.databaseApi.Store(x, y)
}

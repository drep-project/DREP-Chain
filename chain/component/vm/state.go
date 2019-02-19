package vm

import (
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"sync"
	"math/big"
	"errors"
	accountTypes "github.com/drep-project/drep-chain/accounts/types"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/database"
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
	databaseApi *database.DatabaseService
	refund uint64
}

func NewState(databaseService *database.DatabaseService) *State {
	return &State{}
}

func (s *State) CreateContractAccount(callerAddr crypto.CommonAddress, chainId common.ChainIdType, nonce int64) (*accountTypes.Account, error) {
	account, err := accountTypes.NewContractAccount(callerAddr, chainId, nonce)
	if err != nil {
		return nil, err
	}
	return account, s.databaseApi.PutStorage(*account.Address, chainId, account.Storage,true)
}

func (s *State) SubBalance(addr crypto.CommonAddress, chainId common.ChainIdType, amount *big.Int) error {
	balance := s.databaseApi.GetBalance(addr, chainId,true)
	return s.databaseApi.PutBalance(addr, chainId, new(big.Int).Sub(balance, amount), true)
}

func (s *State) AddBalance(addr crypto.CommonAddress, chainId common.ChainIdType, amount *big.Int) error {
	balance := s.databaseApi.GetBalance(addr, chainId,true)
	return s.databaseApi.PutBalance(addr, chainId, new(big.Int).Add(balance, amount),true)
}

func (s *State) GetBalance(addr crypto.CommonAddress, chainId common.ChainIdType) *big.Int {
	return s.databaseApi.GetBalance(addr, chainId,true)
}

func (s *State) SetNonce(addr crypto.CommonAddress, chainId common.ChainIdType, nonce int64) error {
	return s.databaseApi.PutNonce( addr, chainId, nonce,true)
}

func (s *State) GetNonce(addr crypto.CommonAddress, chainId common.ChainIdType) int64 {
	return s.databaseApi.GetNonce(addr, chainId,true)
}

func (s *State) Suicide(addr crypto.CommonAddress, chainId common.ChainIdType) error {
	storage := s.databaseApi.GetStorage(addr, chainId,true)
	storage.Balance = new(big.Int)
	storage.Nonce = 0
	return s.databaseApi.PutStorage(addr, chainId, storage,true)
}

func (s *State) GetByteCode(addr crypto.CommonAddress, chainId common.ChainIdType) crypto.ByteCode {
	return s.databaseApi.GetByteCode(addr, chainId,true)
}

func (s *State) GetCodeSize(addr crypto.CommonAddress, chainId common.ChainIdType) int {
	byteCode := s.GetByteCode(addr, chainId)
	return len(byteCode)
}

func (s *State) GetCodeHash(addr crypto.CommonAddress, chainId common.ChainIdType) crypto.Hash {
	return s.databaseApi.GetCodeHash(addr, chainId,true)
}

func (s *State) SetByteCode(addr crypto.CommonAddress, chainId common.ChainIdType, byteCode crypto.ByteCode) error {
	return s.databaseApi.PutByteCode(addr, chainId, byteCode,true)
}

func (s *State) GetLogs(txHash []byte, chainId common.ChainIdType) []*chainTypes.Log {
	return s.databaseApi.GetLogs(txHash, chainId)
}

func (s *State) AddLog(contractAddr crypto.CommonAddress, chainId common.ChainIdType, txHash, data []byte, topics [][]byte) error {
	log := &chainTypes.Log{
		Address: contractAddr,
		ChainId: chainId,
		TxHash: txHash,
		Data: data,
		Topics: topics,
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

func (s *State) Store(x, y *big.Int, chainId common.ChainIdType) {
	s.databaseApi.Store(x, y)
}
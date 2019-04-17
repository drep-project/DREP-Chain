package vm

import (
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"math/big"
	"sync"
)

var (
	state                   *State
	once                    sync.Once
)

type State struct {
	databaseApi *database.DatabaseService
	refund      uint64
}

func NewState(databaseService *database.DatabaseService) *State {
	return &State{
		databaseApi:databaseService,
	}
}


func (s *State) CreateContractAccount(addr crypto.CommonAddress, byteCode []byte) (*chainTypes.Account, error) {
	account, err := chainTypes.NewContractAccount(addr)
	if err != nil {
		return nil, err
	}
	account.Storage.ByteCode = byteCode
	return account, s.databaseApi.PutStorage(account.Address, account.Storage, true)
}


func (s *State) SubBalance(addr *crypto.CommonAddress, amount *big.Int) error {
	balance := s.databaseApi.GetBalance(addr, true)
	return s.databaseApi.PutBalance(addr, new(big.Int).Sub(balance, amount), true)
}

func (s *State) AddBalance(addr *crypto.CommonAddress, amount *big.Int) error {
	s.databaseApi.AddBalance(addr, amount, true)
	//balance := s.databaseApi.GetBalance(addr, chainId,true)
	//return s.databaseApi.PutBalance(addr, chainId, new(big.Int).Add(balance, amount),true)
	return nil
}

func (s *State) GetBalance(addr *crypto.CommonAddress,) *big.Int {
	return s.databaseApi.GetBalance(addr, true)
}

func (s *State) SetNonce(addr *crypto.CommonAddress, nonce uint64) error {
	return s.databaseApi.PutNonce(addr, nonce, true)
}

func (s *State) GetNonce(addr *crypto.CommonAddress) uint64 {
	return s.databaseApi.GetNonce(addr, true)
}

func (s *State) Suicide(addr *crypto.CommonAddress,) error {
	storage := s.databaseApi.GetStorage(addr, true)
	return s.databaseApi.PutStorage(addr, storage, true)
}

func (s *State) GetByteCode(addr *crypto.CommonAddress,) crypto.ByteCode {
	return s.databaseApi.GetByteCode(addr, true)
}

func (s *State) GetCodeSize(addr crypto.CommonAddress,) int {
	byteCode := s.GetByteCode(&addr)
	return len(byteCode)

}

func (s *State) GetCodeHash(addr crypto.CommonAddress,) crypto.Hash {
	return s.databaseApi.GetCodeHash(&addr, true)
}

func (s *State) SetByteCode(addr *crypto.CommonAddress, byteCode crypto.ByteCode) error {
	return s.databaseApi.PutByteCode(addr, byteCode, true)
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

func (self *State) GetRefund() uint64 {
	return self.refund
}

func (s *State) Load(x *big.Int) []byte {
	return s.databaseApi.Load(x)
}

func (s *State) Store(x, y *big.Int) {
	s.databaseApi.Store(x, y)
}

func (s *State) Exist(contractAddr crypto.CommonAddress) bool {
	storage := s.databaseApi.GetStorage(&contractAddr, true)
	return len(storage.ByteCode) > 0
}

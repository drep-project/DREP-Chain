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
	db     *database.Database
	refund uint64
}

func NewState(database *database.Database) *State {
	return &State{
		db: database,
	}
}


func (s *State) CreateContractAccount(addr crypto.CommonAddress, byteCode []byte) (*chainTypes.Account, error) {
	account, err := chainTypes.NewContractAccount(addr)
	if err != nil {
		return nil, err
	}
	account.Storage.ByteCode = byteCode
	return account, s.db.PutStorage(account.Address, account.Storage)
}


func (s *State) SubBalance(addr *crypto.CommonAddress, amount *big.Int) error {
	balance := s.db.GetBalance(addr)
	return s.db.PutBalance(addr, new(big.Int).Sub(balance, amount))
}

func (s *State) AddBalance(addr *crypto.CommonAddress, amount *big.Int) error {
	return s.db.AddBalance(addr, amount)
}

func (s *State) GetBalance(addr *crypto.CommonAddress,) *big.Int {
	return s.db.GetBalance(addr)
}

func (s *State) SetNonce(addr *crypto.CommonAddress, nonce uint64) error {
	return s.db.PutNonce(addr, nonce)
}

func (s *State) GetNonce(addr *crypto.CommonAddress) uint64 {
	return s.db.GetNonce(addr)
}

func (s *State) Suicide(addr *crypto.CommonAddress,) error {
	storage := s.db.GetStorage(addr)
	return s.db.PutStorage(addr, storage)
}

func (s *State) GetByteCode(addr *crypto.CommonAddress,) crypto.ByteCode {
	return s.db.GetByteCode(addr)
}

func (s *State) GetCodeSize(addr crypto.CommonAddress,) int {
	byteCode := s.GetByteCode(&addr)
	return len(byteCode)

}

func (s *State) GetCodeHash(addr crypto.CommonAddress,) crypto.Hash {
	return s.db.GetCodeHash(&addr)
}

func (s *State) SetByteCode(addr *crypto.CommonAddress, byteCode crypto.ByteCode) error {
	return s.db.PutByteCode(addr, byteCode)
}

func (s *State) GetLogs(txHash []byte,) []*chainTypes.Log {
	return s.db.GetLogs(txHash)
}

func (s *State) AddLog(contractAddr crypto.CommonAddress, txHash, data []byte, topics [][]byte) error {
	log := &chainTypes.Log{
		Address: contractAddr,
		TxHash:  txHash,
		Data:    data,
		Topics:  topics,
	}
	return s.db.AddLog(log)
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
	return s.db.Load(x)
}

func (s *State) Store(x, y *big.Int) {
	s.db.Store(x, y)
}

func (s *State) Exist(contractAddr crypto.CommonAddress) bool {
	storage := s.db.GetStorage(&contractAddr)
	return len(storage.ByteCode) > 0
}

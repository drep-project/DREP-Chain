package vm

import (
	"math/big"
	"sync"

	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/types"
)

var (
	state *State
	once  sync.Once
)

type VMState interface {
	CreateContractAccount(addr crypto.CommonAddress, byteCode []byte) (*types.Account, error)
	SubBalance(addr *crypto.CommonAddress, amount *big.Int) error
	AddBalance(addr *crypto.CommonAddress, amount *big.Int) error
	GetBalance(addr *crypto.CommonAddress) *big.Int
	SetNonce(addr *crypto.CommonAddress, nonce uint64) error
	GetNonce(addr *crypto.CommonAddress) uint64
	Suicide(addr *crypto.CommonAddress) error
	GetByteCode(addr *crypto.CommonAddress) crypto.ByteCode
	GetCodeSize(addr crypto.CommonAddress) int
	GetCodeHash(addr crypto.CommonAddress) crypto.Hash
	SetByteCode(addr *crypto.CommonAddress, byteCode crypto.ByteCode) error
	GetLogs(txHash crypto.Hash) []*types.Log
	AddLog(contractAddr crypto.CommonAddress, txHash crypto.Hash, data []byte, topics []crypto.Hash, blockNumber uint64) error
	AddRefund(gas uint64)
	SubRefund(gas uint64)
	GetRefund() uint64
	Load(x *big.Int) []byte
	Store(x, y *big.Int)
	Exist(contractAddr crypto.CommonAddress) bool
}

type State struct {
	db     *database.Database
	refund uint64
}

func NewState(database *database.Database) *State {
	return &State{
		db: database,
	}
}

func (s *State) CreateContractAccount(addr crypto.CommonAddress, byteCode []byte) (*types.Account, error) {
	account, err := types.NewContractAccount(addr)
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

func (s *State) GetBalance(addr *crypto.CommonAddress) *big.Int {
	return s.db.GetBalance(addr)
}

func (s *State) SetNonce(addr *crypto.CommonAddress, nonce uint64) error {
	return s.db.PutNonce(addr, nonce)
}

func (s *State) GetNonce(addr *crypto.CommonAddress) uint64 {
	return s.db.GetNonce(addr)
}

func (s *State) Suicide(addr *crypto.CommonAddress) error {
	storage := s.db.GetStorage(addr)
	return s.db.PutStorage(addr, storage)
}

func (s *State) GetByteCode(addr *crypto.CommonAddress) crypto.ByteCode {
	return s.db.GetByteCode(addr)
}

func (s *State) GetCodeSize(addr crypto.CommonAddress) int {
	byteCode := s.GetByteCode(&addr)
	return len(byteCode)

}

func (s *State) GetCodeHash(addr crypto.CommonAddress) crypto.Hash {
	return s.db.GetCodeHash(&addr)
}

func (s *State) SetByteCode(addr *crypto.CommonAddress, byteCode crypto.ByteCode) error {
	return s.db.PutByteCode(addr, byteCode)
}

func (s *State) GetLogs(txHash crypto.Hash) []*types.Log {
	return s.db.GetLogs(txHash)
}

func (s *State) AddLog(contractAddr crypto.CommonAddress, txHash crypto.Hash, data []byte, topics []crypto.Hash, blockNumber uint64) error {
	log := &types.Log{
		Address: contractAddr,
		TxHash:  txHash,
		Data:    data,
		Topics:  topics,
		Height:  blockNumber,
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

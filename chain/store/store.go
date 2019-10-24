package store

import (
	"github.com/drep-project/drep-chain/common/trie"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/database/dbinterface"
	dlog "github.com/drep-project/drep-chain/pkgs/log"
	"math/big"
)

const (
	MODULENAME = "store"
)

var (
	log = dlog.EnsureLogger(MODULENAME)
)

type StoreInterface interface {
	DeleteStorage(addr *crypto.CommonAddress) error

	GetStorageAlias(addr *crypto.CommonAddress) string
	AliasGet(alias string) (*crypto.CommonAddress, error)
	AliasExist(alias string) bool
	AliasSet(addr *crypto.CommonAddress, alias string) (err error)

	GetBalance(addr *crypto.CommonAddress, height uint64) *big.Int
	PutBalance(addr *crypto.CommonAddress, height uint64, balance *big.Int) error
	AddBalance(addr *crypto.CommonAddress, height uint64, amount *big.Int) error
	SubBalance(addr *crypto.CommonAddress, height uint64, amount *big.Int) error

	GetNonce(addr *crypto.CommonAddress) uint64
	PutNonce(addr *crypto.CommonAddress, nonce uint64) error

	GetByteCode(addr *crypto.CommonAddress) []byte
	GetCodeHash(addr *crypto.CommonAddress) crypto.Hash
	PutByteCode(addr *crypto.CommonAddress, byteCode []byte) error

	GetReputation(addr *crypto.CommonAddress) *big.Int
	GetStateRoot() []byte
	RecoverTrie(root []byte) bool

	TrieDB() *trie.Database
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Commit()

	CopyState() *database.SnapShot
	RevertState(shot *database.SnapShot)

	Empty(addr *crypto.CommonAddress) bool

	//pos
	GetCandidateAddrs() ([]crypto.CommonAddress, error)
	GetVoteCreditCount(addr *crypto.CommonAddress) *big.Int
	CancelVoteCredit(fromAddr, toAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) error
	VoteCredit(addresses *crypto.CommonAddress, to *crypto.CommonAddress, addBalance *big.Int, height uint64) error
	CandidateCredit(addresses *crypto.CommonAddress, addBalance *big.Int, data []byte, height uint64) error
	CancelCandidateCredit(fromAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) error
	GetCandidateData(addr *crypto.CommonAddress) ([]byte, error)
	AddCandidateAddr(addr *crypto.CommonAddress) error
}

type Store struct {
	stake   *trieStakeStore
	account *trieAccountStore
	db      *StoreDB
}

func (s Store) AddCandidateAddr(addr *crypto.CommonAddress) error {
	return s.stake.AddCandidateAddr(addr)
}

func (s Store) CancelCandidateCredit(fromAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) error {
	return s.stake.CancelCandidateCredit(fromAddr, cancelBalance, height)
}

func (s Store) GetCandidateData(addr *crypto.CommonAddress) ([]byte, error) {
	return s.stake.GetCandidateData(addr)
}

func (s Store) GetCandidateAddrs() ([]crypto.CommonAddress, error) {
	return s.stake.GetCandidateAddrs()
}

func (s Store) GetVoteCreditCount(addr *crypto.CommonAddress) *big.Int {
	return s.stake.GetCreditCount(addr)
}

func (s Store) Empty(addr *crypto.CommonAddress) bool {
	return s.Empty(addr)
}

func (s Store) GetStorageAlias(addr *crypto.CommonAddress) string {
	return s.account.GetStorageAlias(addr)
}

func (s Store) DeleteStorage(addr *crypto.CommonAddress) error {
	return s.account.DeleteStorage(addr)
}

func (s Store) AliasGet(alias string) (*crypto.CommonAddress, error) {
	return s.account.AliasGet(alias)
}

func (s Store) AliasExist(alias string) bool {
	return s.account.AliasExist(alias)
}

func (s Store) AddBalance(addr *crypto.CommonAddress, height uint64, amount *big.Int) error {
	voteCredit, err := s.stake.CancelCreditToBalance(addr, height)
	if err != nil {
		return err
	}
	return s.account.AddBalance(addr, amount.Add(amount, voteCredit))
}

func (s Store) SubBalance(addr *crypto.CommonAddress, height uint64, amount *big.Int) error {
	voteCredit, err := s.stake.CancelCreditToBalance(addr, height)
	if err != nil {
		return err
	}

	err = s.account.AddBalance(addr, voteCredit)
	if err != nil {
		return err
	}

	return s.account.SubBalance(addr, amount)
}

func (s Store) GetBalance(addr *crypto.CommonAddress, height uint64) *big.Int {
	return new(big.Int).Add(s.stake.GetCancelCreditForBalance(addr, height), s.account.GetBalance(addr))
}

func (s Store) PutBalance(addr *crypto.CommonAddress, height uint64, balance *big.Int) error {
	voteCredit, err := s.stake.CancelCreditToBalance(addr, height)
	if err != nil {
		return err
	}

	err = s.account.AddBalance(addr, voteCredit)
	if err != nil {
		return err
	}
	return s.account.PutBalance(addr, balance)
}

func (s Store) GetNonce(addr *crypto.CommonAddress) uint64 {
	return s.account.GetNonce(addr)
}

func (s Store) PutNonce(addr *crypto.CommonAddress, nonce uint64) error {
	return s.account.PutNonce(addr, nonce)
}

func (s Store) GetByteCode(addr *crypto.CommonAddress) []byte {
	return s.account.GetByteCode(addr)
}

func (s Store) GetCodeHash(addr *crypto.CommonAddress) crypto.Hash {
	return s.account.GetCodeHash(addr)
}

func (s Store) PutByteCode(addr *crypto.CommonAddress, byteCode []byte) error {
	return s.account.PutByteCode(addr, byteCode)
}

func (s Store) GetReputation(addr *crypto.CommonAddress) *big.Int {
	return s.account.GetReputation(addr)
}

func (s Store) AliasSet(addr *crypto.CommonAddress, alias string) (err error) {
	return s.account.AliasSet(addr, alias)
}

func (s Store) CancelVoteCredit(fromAddr, toAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) error {
	return s.stake.CancelVoteCredit(fromAddr, toAddr, cancelBalance, height)
}

func (s Store) VoteCredit(fromAddr *crypto.CommonAddress, to *crypto.CommonAddress, addBalance *big.Int, height uint64) error {
	return s.stake.VoteCredit(fromAddr, to, addBalance, height)
}

func (s Store) CandidateCredit(fromAddr *crypto.CommonAddress, addBalance *big.Int, data []byte, height uint64) error {
	return s.stake.CandidateCredit(fromAddr, addBalance, data, height)
}
func (s Store) TrieDB() *trie.Database {
	return s.account.TrieDB()
}

func (s Store) Get(key []byte) ([]byte, error) {
	return s.db.Get(key)
}

func (s Store) Put(key []byte, value []byte) error {
	return s.db.Put(key, value)
}

func TrieStoreFromStore(diskDB dbinterface.KeyValueStore, stateRoot []byte) (StoreInterface, error) {
	db := NewStoreDB(diskDB, nil, nil, trie.NewDatabaseWithCache(diskDB, 0))

	store := &Store{
		stake:   NewStakeStorage(db),
		account: NewTrieAccoutStore(db),
		db:      db,
	}

	if !store.RecoverTrie(stateRoot) {
		return nil, ErrRecoverRoot
	}

	return store, nil
}

func (s *Store) RevertState(shot *database.SnapShot) {
	s.db.RevertState(shot)
}

func (s Store) CopyState() *database.SnapShot {
	return s.db.CopyState()
}

func (s Store) Commit() {
	s.db.Flush()
}

func (s Store) GetStateRoot() []byte {
	s.db.Flush()
	return s.db.getStateRoot()
}

func (s *Store) RecoverTrie(root []byte) bool {
	return s.db.RecoverTrie(root)
}

func (s *Store) GetCreditDetails(addr *crypto.CommonAddress) map[crypto.CommonAddress]big.Int {
	return s.stake.GetCreditDetails(addr)
}

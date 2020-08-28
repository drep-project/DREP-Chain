package store

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/drep-project/DREP-Chain/common/trie"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/database"
	"github.com/drep-project/DREP-Chain/database/dbinterface"
	"github.com/drep-project/DREP-Chain/params"
	dlog "github.com/drep-project/DREP-Chain/pkgs/log"
	"github.com/drep-project/DREP-Chain/types"
)

const (
	MODULENAME     = "store"
	ChangeInterval = "changeInterval"
)

var (
	log = dlog.EnsureLogger(MODULENAME)
)

type StoreInterface interface {
	DeleteStorage(addr *crypto.CommonAddress) error

	GetStorageAlias(addr *crypto.CommonAddress) string
	AliasGet(alias string) (*crypto.CommonAddress, error)
	AliasExist(alias string) bool
	AliasSet(addr *crypto.CommonAddress, alias string, height uint64) (err error)

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
	GetChangeInterval() (uint64, error)

	//pos
	GetCandidateAddrs() ([]crypto.CommonAddress, error)
	GetVoteCreditCount(addr *crypto.CommonAddress) *big.Int
	CancelVoteCredit(fromAddr, toAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) (*types.CancelCreditDetail, error)
	VoteCredit(addresses *crypto.CommonAddress, to *crypto.CommonAddress, addBalance *big.Int, height uint64) error
	CandidateCredit(addresses *crypto.CommonAddress, addBalance *big.Int, data []byte, height uint64) error
	CancelCandidateCredit(fromAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) (*types.CancelCreditDetail, error)
	GetCandidateData(addr *crypto.CommonAddress) ([]byte, error)
	AddCandidateAddr(addr *crypto.CommonAddress) error
	GetCreditDetails(addr *crypto.CommonAddress) map[crypto.CommonAddress]big.Int
}

type Store struct {
	stake   *trieStakeStore
	account *trieAccountStore
	db      *StoreDB
}

//Show interest rates in 3 months, 3-6 months, 6-12 months and more
func GetInterestRate() (threeMonth, sixMonth, oneYear, moreOneYear uint64) {
	var rate uint64 = 0
	rate = interestRate * 8
	bigDiff := new(big.Int).SetUint64(threeMonthHeight * 100)
	//Less than 3 months
	threeMonth = bigDiff.Div(bigDiff, new(big.Int).SetUint64(rate)).Uint64()
	threeMonth = threeMonth * (12 / 3) //Annualized interest rate

	rate = interestRate * 4
	bigDiff = new(big.Int).SetUint64(sixMonthHeight * 100)
	//3 to 6 months
	sixMonth = bigDiff.Div(bigDiff, new(big.Int).SetUint64(rate)).Uint64()
	sixMonth = sixMonth * (12 / 6) //Annualized interest rate

	rate = interestRate * 2
	bigDiff = new(big.Int).SetUint64(oneYearHeight * 100)
	//6 to 12 months
	oneYear = bigDiff.Div(bigDiff, new(big.Int).SetUint64(rate)).Uint64()
	oneYear = oneYear * (12 / 12) //Annualized interest rate

	rate = interestRate
	bigDiff = new(big.Int).SetUint64(oneYearHeight * 100)
	//More than 12 months
	moreOneYear = bigDiff.Div(bigDiff, new(big.Int).SetUint64(rate)).Uint64()

	return
}

func (s Store) AddCandidateAddr(addr *crypto.CommonAddress) error {
	return s.stake.AddCandidateAddr(addr)
}

func (s Store) GetChangeInterval() (uint64, error) {
	value, err := s.db.store.Get([]byte(ChangeInterval))
	if err != nil {
		log.Error("CancelCandidateCredit get change interval ", "err", err)
		return 0, err
	}
	var changeInterval uint64
	buf := bytes.NewBuffer(value)
	err = binary.Read(buf, binary.BigEndian, &changeInterval)
	if err != nil {
		log.Error("CancelCandidateCredit parse change interval ", "err", err)
		return 0, err
	}

	return changeInterval * 100, nil //after 10000 block,refund to account

}
func (s Store) CancelCandidateCredit(fromAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) (*types.CancelCreditDetail, error) {
	ci, err := s.GetChangeInterval()
	if err != nil {
		return nil, err
	}
	return s.stake.CancelCandidateCredit(fromAddr, cancelBalance, height, ci)
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
	ci, err := s.GetChangeInterval()
	if err != nil {
		return err
	}

	voteCredit, err := s.stake.CancelCreditToBalance(addr, height, ci)
	if err != nil {
		return err
	}
	return s.account.AddBalance(addr, amount.Add(amount, voteCredit))
}

func (s Store) SubBalance(addr *crypto.CommonAddress, height uint64, amount *big.Int) error {
	ci, err := s.GetChangeInterval()
	if err != nil {
		return err
	}
	voteCredit, err := s.stake.CancelCreditToBalance(addr, height, ci)
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
	ci, err := s.GetChangeInterval()
	if err != nil {
		return nil
	}
	return new(big.Int).Add(s.stake.GetCancelCreditForBalance(addr, height, ci), s.account.GetBalance(addr))
}

func (s Store) PutBalance(addr *crypto.CommonAddress, height uint64, balance *big.Int) error {
	if height != 0 {
		ci, err := s.GetChangeInterval()
		if err != nil {
			return err
		}
		voteCredit, err := s.stake.CancelCreditToBalance(addr, height, ci)
		if err != nil {
			return err
		}

		err = s.account.AddBalance(addr, voteCredit)
		if err != nil {
			return err
		}
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

func (s Store) AliasSet(addr *crypto.CommonAddress, alias string, height uint64) (err error) {
	drepFee, err := types.CheckAlias([]byte(alias))
	if err != nil {
		return
	}
	//minus alias fee from from account
	originBalance := s.GetBalance(addr, height)
	leftBalance := originBalance.Sub(originBalance, drepFee)
	if leftBalance.Sign() < 0 {
		return errors.New("set alias ,not enough balance")
	}
	err = s.PutBalance(addr, height, leftBalance)
	if err != nil {
		return
	}
	// put alias fee to hole address
	zeroAddressBalance := s.GetBalance(&params.HoleAddress, height)
	zeroAddressBalance = zeroAddressBalance.Add(zeroAddressBalance, drepFee)
	err = s.PutBalance(&params.HoleAddress, height, zeroAddressBalance)
	if err != nil {
		return
	}
	return s.account.AliasSet(addr, alias)
}

func (s Store) CancelVoteCredit(fromAddr, toAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) (*types.CancelCreditDetail, error) {
	ci, err := s.GetChangeInterval()
	if err != nil {
		return nil, err
	}
	return s.stake.CancelVoteCredit(fromAddr, toAddr, cancelBalance, height, ci)
}

func (s Store) VoteCredit(fromAddr *crypto.CommonAddress, to *crypto.CommonAddress, addBalance *big.Int, height uint64) error {
	return s.stake.VoteCredit(fromAddr, to, addBalance, height)
}

func (s Store) CandidateCredit(fromAddr *crypto.CommonAddress, addBalance *big.Int, data []byte, height uint64) error {
	cd := types.CandidateData{}
	if err := cd.Unmarshal(data); nil != err {
		return err
	}
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
		stake:   newStakeStorage(db),
		account: newTrieAccoutStore(db),
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

package bft

import (
	"crypto/rand"
	"github.com/drep-project/DREP-Chain/common/trie"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/database"
	"github.com/drep-project/DREP-Chain/types"
	"math/big"
	"testing"
)

type fakeStore struct {
}

func (fakeStore) DeleteStorage(addr *crypto.CommonAddress) error {
	panic("implement me")
}

func (fakeStore) GetStorageAlias(addr *crypto.CommonAddress) string {
	panic("implement me")
}

func (fakeStore) AliasGet(alias string) (*crypto.CommonAddress, error) {
	panic("implement me")
}

func (fakeStore) AliasExist(alias string) bool {
	panic("implement me")
}

func (fakeStore) AliasSet(addr *crypto.CommonAddress, alias string) (err error) {
	panic("implement me")
}

func (fakeStore) GetBalance(addr *crypto.CommonAddress, height uint64) *big.Int {
	panic("implement me")
}

func (fakeStore) PutBalance(addr *crypto.CommonAddress, height uint64, balance *big.Int) error {
	panic("implement me")
}

func (fakeStore) AddBalance(addr *crypto.CommonAddress, height uint64, amount *big.Int) error {
	//fmt.Println("add balance:", addr.String(), height, amount)
	return nil
}

func (fakeStore) SubBalance(addr *crypto.CommonAddress, height uint64, amount *big.Int) error {
	panic("implement me")
}

func (fakeStore) GetNonce(addr *crypto.CommonAddress) uint64 {
	panic("implement me")
}

func (fakeStore) PutNonce(addr *crypto.CommonAddress, nonce uint64) error {
	panic("implement me")
}

func (fakeStore) GetByteCode(addr *crypto.CommonAddress) []byte {
	panic("implement me")
}

func (fakeStore) GetCodeHash(addr *crypto.CommonAddress) crypto.Hash {
	panic("implement me")
}

func (fakeStore) PutByteCode(addr *crypto.CommonAddress, byteCode []byte) error {
	panic("implement me")
}

func (fakeStore) GetReputation(addr *crypto.CommonAddress) *big.Int {
	panic("implement me")
}

func (fakeStore) GetStateRoot() []byte {
	panic("implement me")
}

func (fakeStore) RecoverTrie(root []byte) bool {
	panic("implement me")
}

func (fakeStore) TrieDB() *trie.Database {
	panic("implement me")
}

func (fakeStore) Get(key []byte) ([]byte, error) {
	panic("implement me")
}

func (fakeStore) Put(key []byte, value []byte) error {
	panic("implement me")
}

func (fakeStore) Commit() {
	panic("implement me")
}

func (fakeStore) CopyState() *database.SnapShot {
	panic("implement me")
}

func (fakeStore) RevertState(shot *database.SnapShot) {
	panic("implement me")
}

func (fakeStore) Empty(addr *crypto.CommonAddress) bool {
	panic("implement me")
}

func (fakeStore) GetChangeInterval() (uint64, error) {
	panic("implement me")
}

func (fakeStore) GetCandidateAddrs() ([]crypto.CommonAddress, error) {
	panic("implement me")
}

func (fakeStore) GetVoteCreditCount(addr *crypto.CommonAddress) *big.Int {
	panic("implement me")
}

func (fakeStore) CancelVoteCredit(fromAddr, toAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) (*types.CancelCreditDetail, error) {
	panic("implement me")
}

func (fakeStore) VoteCredit(addresses *crypto.CommonAddress, to *crypto.CommonAddress, addBalance *big.Int, height uint64) error {
	panic("implement me")
}

func (fakeStore) CandidateCredit(addresses *crypto.CommonAddress, addBalance *big.Int, data []byte, height uint64) error {
	panic("implement me")
}

func (fakeStore) CancelCandidateCredit(fromAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) (*types.CancelCreditDetail, error) {
	panic("implement me")
}

func (fakeStore) GetCandidateData(addr *crypto.CommonAddress) ([]byte, error) {
	panic("implement me")
}

func (fakeStore) AddCandidateAddr(addr *crypto.CommonAddress) error {
	panic("implement me")
}

func (fakeStore) GetCreditDetails(addr *crypto.CommonAddress) map[crypto.CommonAddress]big.Int {
	m := make(map[crypto.CommonAddress]big.Int)

	for i := 1; i < 6; i++ {
		pri, _ := crypto.GenerateKey(rand.Reader)
		backbone := crypto.PubkeyToAddress(pri.PubKey())
		m[backbone] = *new(big.Int).SetInt64(int64(i * 10))
	}
	return m
}

func TestRewardCalculator_AccumulateRewards(t *testing.T) {
	fs := fakeStore{}
	ms := MultiSignature{}

	ps := make(ProducerSet, 0, 3)
	for i := 0; i < 3; i++ {
		pk, _ := crypto.GenerateKey(rand.Reader)
		ps = append(ps, Producer{Pubkey: pk.PubKey()})
	}

	nc := NewRewardCalculator(fs, &ms, ps, new(big.Int).SetInt64(100), 100)
	err := nc.AccumulateRewards(120)
	if err != nil {
		panic("reward errrrrrrrrr")
	}
}

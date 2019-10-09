package pos

import (
	"crypto/rand"
	"fmt"
	"github.com/drep-project/drep-chain/chain/store"
	"github.com/drep-project/drep-chain/common/trie"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"math/big"
	"testing"
)

type StoreFake struct {
	m map[crypto.CommonAddress]struct{}
}

func (StoreFake) DeleteStorage(addr *crypto.CommonAddress) error {
	panic("implement me")
}

func (StoreFake) GetStorageAlias(addr *crypto.CommonAddress) string {
	panic("implement me")
}

func (StoreFake) AliasGet(alias string) (*crypto.CommonAddress, error) {
	panic("implement me")
}

func (StoreFake) AliasExist(alias string) bool {
	panic("implement me")
}

func (StoreFake) GetBalance(addr *crypto.CommonAddress, height uint64) *big.Int {
	panic("implement me")
}

func (StoreFake) PutBalance(addr *crypto.CommonAddress, height uint64, balance *big.Int) error {
	panic("implement me")
}

func (StoreFake) AddBalance(addr *crypto.CommonAddress, height uint64, amount *big.Int) error {
	panic("implement me")
}

func (StoreFake) SubBalance(addr *crypto.CommonAddress, height uint64, amount *big.Int) error {
	panic("implement me")
}

func (StoreFake) GetNonce(addr *crypto.CommonAddress) uint64 {
	panic("implement me")
}

func (StoreFake) PutNonce(addr *crypto.CommonAddress, nonce uint64) error {
	panic("implement me")
}

func (StoreFake) GetByteCode(addr *crypto.CommonAddress) []byte {
	panic("implement me")
}

func (StoreFake) GetCodeHash(addr *crypto.CommonAddress) crypto.Hash {
	panic("implement me")
}

func (StoreFake) PutByteCode(addr *crypto.CommonAddress, byteCode []byte) error {
	panic("implement me")
}

func (StoreFake) GetReputation(addr *crypto.CommonAddress) *big.Int {
	panic("implement me")
}

func (StoreFake) GetStateRoot() []byte {
	panic("implement me")
}

func (StoreFake) RecoverTrie(root []byte) bool {
	panic("implement me")
}

func (StoreFake) AliasSet(addr *crypto.CommonAddress, alias string) (err error) {
	panic("implement me")
}

func (StoreFake) CancelVoteCredit(fromAddr, toAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) error {
	panic("implement me")
}

func (StoreFake) TrieDB() *trie.Database {
	panic("implement me")
}

func (StoreFake) Get(key []byte) ([]byte, error) {
	panic("implement me")
}

func (StoreFake) Put(key []byte, value []byte) error {
	panic("implement me")
}

func (StoreFake) VoteCredit(addresses *crypto.CommonAddress, to *crypto.CommonAddress, addBalance *big.Int) error {
	panic("implement me")
}

func (StoreFake) CopyState() *database.SnapShot {
	panic("implement me")
}

func (StoreFake) RevertState(shot *database.SnapShot) {
	panic("implement me")
}

func (StoreFake) Empty(addr *crypto.CommonAddress) bool {
	panic("implement me")
}

func (s StoreFake) GetCandidateAddrs() (map[crypto.CommonAddress]struct{}, error) {
	return s.m, nil
}

func (s StoreFake) GetVoteCredit(addr *crypto.CommonAddress) *big.Int {
	if _, ok := s.m[*addr]; ok {
		rd, _ := rand.Int(rand.Reader, new(big.Int).SetUint64(1000000))
		fmt.Println(addr.String(),rd)
		return rd
	}
	return &big.Int{}
}

func NewStoreFake() *StoreFake {
	s := &StoreFake{m: make(map[crypto.CommonAddress]struct{})}

	for i := 0; i < 20; i++ {
		privKey, _ := crypto.GenerateKey(rand.Reader)
		addr := crypto.PubkeyToAddress(privKey.PubKey())
		s.m[addr] = struct{}{}
	}

	return s
}

func TestGetCandidates(t *testing.T) {
	var si store.StoreInterface
	si = NewStoreFake()
	addrs := GetCandidates(si)
	for _,a := range addrs{
		fmt.Println(a.String())
	}
}

func TestUpdateCandidateStake(t *testing.T) {

}

func TestAddCandidate(t *testing.T) {

}

func TestCancelCandidate(t *testing.T) {

}

func TestRecoverCandidates(t *testing.T) {

}

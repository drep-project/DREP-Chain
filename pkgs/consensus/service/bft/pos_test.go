package bft

import (
	"crypto/rand"
	"fmt"
	"github.com/drep-project/DREP-Chain/chain/store"
	"github.com/drep-project/DREP-Chain/common/trie"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/database"
	"github.com/drep-project/DREP-Chain/types"
	"math/big"
	"testing"
)

var topN int = 18

type StoreFake struct {
	m map[crypto.CommonAddress]struct{}
}

func (s StoreFake) Commit() {
	panic("implement me")
}

func (s StoreFake) GetChangeInterval() (uint64, error) {
	panic("implement me")
}

func (s StoreFake) GetCandidateAddrs() ([]crypto.CommonAddress, error) {
	addrs := make([]crypto.CommonAddress, 0)
	for k, _ := range s.m {
		addrs = append(addrs, k)
	}
	return addrs, nil
}

func (s StoreFake) CancelVoteCredit(fromAddr, toAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) (*types.IntersetDetail, error) {
	panic("implement me")
}

func (s StoreFake) VoteCredit(addresses *crypto.CommonAddress, to *crypto.CommonAddress, addBalance *big.Int, height uint64) error {
	panic("implement me")
}

func (s StoreFake) CandidateCredit(addresses *crypto.CommonAddress, addBalance *big.Int, data []byte, height uint64) error {
	panic("implement me")
}

func (s StoreFake) CancelCandidateCredit(fromAddr *crypto.CommonAddress, cancelBalance *big.Int, height uint64) (*types.IntersetDetail, error) {
	panic("implement me")
}

func (s StoreFake) AddCandidateAddr(addr *crypto.CommonAddress) error {
	panic("implement me")
}

func (s StoreFake) GetCandidateData(addr *crypto.CommonAddress) ([]byte, error) {
	//pk, _ := crypto.GenerateKey(rand.Reader)

	cd := &types.CandidateData{}
	cd.Pubkey = pubkeys[*addr]
	cd.Node = "enode://e77d64fecbb1c7e78231507fdd58c963cdc1e0ed0bec29b5a65de32b992d596f@149.129.172.91:44444"

	return cd.Marshal()
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

func (StoreFake) TrieDB() *trie.Database {
	panic("implement me")
}

func (StoreFake) Get(key []byte) ([]byte, error) {
	panic("implement me")
}

func (StoreFake) Put(key []byte, value []byte) error {
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

var getNum int = 0

func (s StoreFake) GetVoteCreditCount(addr *crypto.CommonAddress) *big.Int {

	if getNum < topN/2 {
		getNum++
		if _, ok := s.m[*addr]; ok {
			//rd, _ := rand.Int(rand.Reader, new(big.Int).SetUint64(1000000))
			rd := new(big.Int).SetUint64(1000000)
			fmt.Println(addr.String(), rd)
			return rd
		}
	} else {
		if _, ok := s.m[*addr]; ok {
			rd, _ := rand.Int(rand.Reader, new(big.Int).SetUint64(1000000))
			//rd := new(big.Int).SetUint64(10000)
			fmt.Println(addr.String(), rd)
			return rd
		}
	}

	return &big.Int{}
}

var pubkeys map[crypto.CommonAddress]*secp256k1.PublicKey

func NewStoreFake() *StoreFake {
	s := &StoreFake{m: make(map[crypto.CommonAddress]struct{})}
	pubkeys = make(map[crypto.CommonAddress]*secp256k1.PublicKey)

	for i := 0; i < 20; i++ {
		privKey, _ := crypto.GenerateKey(rand.Reader)
		addr := crypto.PubkeyToAddress(privKey.PubKey())
		s.m[addr] = struct{}{}
		pubkeys[addr] = privKey.PubKey()
	}

	return s
}

func TestGetCandidates(t *testing.T) {
	var si store.StoreInterface
	si = NewStoreFake()
	addrs := GetCandidates(si, topN)
	for i, data := range addrs {
		fmt.Println(i, data.Address().String(), data.Node)
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

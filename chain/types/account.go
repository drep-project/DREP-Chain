package types

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"math/big"
)

var (
	DrepMark   = []byte("Drep Coin Seed")
	KeyBitSize = 256 >> 3
)

type KeyWeight struct {
	Key    secp256k1.PublicKey
	Weight uint16
}

func (key KeyWeight) Compare(kw KeyWeight) bool {
	if !key.Key.IsEqual(&kw.Key) {
		return false
	}
	if key.Weight != kw.Weight {
		return false
	}
	return true
}

type Authority struct {
	Threshold uint32                  `json:"threshold"`
	Keys      []KeyWeight             `json:"keys"`
}

func NewAuthority(k secp256k1.PublicKey) (a Authority) {
	a.Threshold = 1
	a.Keys = append(a.Keys, KeyWeight{k, 1})
	return a
}

type Storage struct {
	Name   		string
	Authority	Authority

	ChainId    app.ChainIdType
	ChainCode  []byte
	Balance    	*big.Int
	Nonce      	int64
	Reputation 	*big.Int
	//contract
	ByteCode   	crypto.ByteCode
	CodeHash   	crypto.Hash
	//bios
	Miner 		map[string]*secp256k1.PublicKey
}

func NewStorage(name string, chainId app.ChainIdType, chainCode []byte, authority Authority) *Storage {
	return &Storage{
		Name: name,
		ChainId  : chainId,
		ChainCode: chainCode,
		Nonce:0,
		Balance : new(big.Int),
		Authority: authority,
		Miner: map[string]*secp256k1.PublicKey{},
	}
}

type Account struct {
	Name 	string
	Storage *Storage
}
func RandomAccount() (*secp256k1.PrivateKey, []byte){
	uni, err := common.GenUnique()
	if err != nil {
		return nil, nil
	}
	h := common.HmAC(uni, DrepMark)
	prvKey, _ := secp256k1.PrivKeyFromBytes(h[:KeyBitSize])
	chainCode := h[KeyBitSize:]
	return prvKey, chainCode
}

func NewAccount(name string, chainId app.ChainIdType, chainCode []byte, pubkey secp256k1.PublicKey) *Account{
	return &Account{
		Name: name,
		Storage: NewStorage(name, chainId, chainCode, NewAuthority(pubkey)),
	}
}

func (account *Account) Derive(privKey *secp256k1.PrivateKey) (*Account, *secp256k1.PrivateKey){
	pid := new(big.Int).SetBytes( account.Storage.ChainCode)
	cid := new(big.Int).SetBytes( account.Storage.ChainId[:])
	chainCode := new(big.Int).Xor(pid, cid).Bytes()

	h := common.HmAC(chainCode, privKey.Serialize())
	newPrvKey, _ := secp256k1.PrivKeyFromBytes(h[:KeyBitSize])
	chainCode = h[KeyBitSize:]

	return &Account{
		Name: account.Name,
		Storage: &Storage{
			Name: account.Name,
			ChainId  : account.Storage.ChainId,
			ChainCode: chainCode,
		},
	}, newPrvKey
}

func NewContractAccount(contractName string, chainId app.ChainIdType) (*Account, error) {
	account := &Account{
		Name: 	 contractName,
		Storage: &Storage{
			Name: contractName,
			ChainId  : chainId,
			ChainCode: []byte{},    //contract address cannot Derive
		},
	}
	return account, nil
}

package types

import (
	"math/big"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
)

var (
	DrepMark   = []byte("Drep Coin Seed")
	KeyBitSize = 256 >> 3
)

type Node struct {
	Address    *crypto.CommonAddress
	PrivateKey *secp256k1.PrivateKey
	ChainId    ChainIdType
	ChainCode  []byte
}

func NewNode(parent *Node, chainId ChainIdType) *Node {
	var (
		prvKey    *secp256k1.PrivateKey
		chainCode []byte
	)

	IsRoot := parent == nil
	if IsRoot {
		uni, err := common.GenUnique()
		if err != nil {
			return nil
		}
		h := common.HmAC(uni, DrepMark)
		prvKey, _ = secp256k1.PrivKeyFromBytes(h[:KeyBitSize])
		chainCode = h[KeyBitSize:]
	} else {
		pid := new(big.Int).SetBytes(parent.ChainCode)
		cid := new(big.Int).SetBytes(chainId.Bytes())
		chainCode := new(big.Int).Xor(pid, cid).Bytes()

		h := common.HmAC(chainCode, parent.PrivateKey.Serialize())
		prvKey, _ = secp256k1.PrivKeyFromBytes(h[:KeyBitSize])
		chainCode = h[KeyBitSize:]
	}
	address := crypto.PubKey2Address(prvKey.PubKey())
	return &Node{
		Address:    &address,
		PrivateKey: prvKey,
		ChainId:    chainId,
		ChainCode:  chainCode,
	}
}

type Storage struct {
	Balance    big.Int
	Nonce      uint64
	ByteCode   crypto.ByteCode
	CodeHash   crypto.Hash
	Reputation *big.Int
	Alias      string
}

func NewStorage() *Storage {
	storage := &Storage{}
	storage.Nonce = 0
	return storage
}

type Account struct {
	Address *crypto.CommonAddress
	Node    *Node
	Storage *Storage
}

func (account *Account) Sign(hash []byte) ([]byte, error) {
	return crypto.Sign(hash, account.Node.PrivateKey)
}

func NewNormalAccount(parent *Node, chainId ChainIdType) (*Account, error) {
	/*IsRoot := chainId == RootChain
	if !IsRoot && parent == nil {
		return nil, errors.New("missing parent account")
	}*/
	node := NewNode(parent, chainId)
	address := node.Address
	storage := NewStorage()
	account := &Account{
		Address: address,
		Node:    node,
		Storage: storage,
	}
	return account, nil
}

func NewContractAccount(address crypto.CommonAddress) (*Account, error) {
	storage := NewStorage()
	account := &Account{
		Address: &address,
		Node:    &Node{},
		Storage: storage,
	}
	return account, nil
}

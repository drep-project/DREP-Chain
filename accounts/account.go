package accounts

import (
	"math/big"
	"BlockChainTest/mycrypto"
	"errors"
)

var (
	DrepMark   = []byte("Drep Coin Seed")
	KeyBitSize = 256 >> 3
	DefaultRep = new(big.Int).SetInt64(10000)
)

type Node struct {
	PrvKey    *mycrypto.PrivateKey
	ChainId   int64
	ChainCode []byte
}

func NewNode(parent *Node, chainId int64) *Node {
	var (
		prvKey *mycrypto.PrivateKey
		chainCode []byte
	)

	IsRoot := parent == nil
	if IsRoot {
		uni, err := genUnique()
		if err != nil {
			return nil
		}
		h := hmAC(uni, DrepMark)
		prvKey = genPrvKey(h[:KeyBitSize])
		chainCode = h[KeyBitSize:]
	} else {
		pid := new(big.Int).SetBytes(parent.ChainCode)
		cid := new(big.Int).SetInt64(int64(chainId))
		msg := new(big.Int).Xor(pid, cid).Bytes()
		h := hmAC(msg, parent.PrvKey.Prv)
		prvKey = genPrvKey(h[:KeyBitSize])
	}

	return &Node{
		PrvKey: prvKey,
		ChainId: chainId,
		ChainCode: chainCode,
	}
}

func (node *Node) Address() CommonAddress {
	return PubKey2Address(node.PrvKey.PubKey)
}

type Storage struct {
	Balance   *big.Int
	Nonce     int64
	ByteCode  ByteCode
	CodeHash  Hash
}

func NewStorage() *Storage {
	storage := &Storage{}
	storage.Balance = new(big.Int)
	storage.Nonce = 0
	return storage
}

type Account struct {
	Address       CommonAddress
	Node          *Node
	Storage       *Storage
}

func NewNormalAccount(parent *Node, chainId int64) (*Account, error) {
	IsRoot := chainId == RootChainID
	if !IsRoot && parent == nil {
		return nil, errors.New("missing parent account")
	}
	node := NewNode(parent, chainId)
	address := node.Address()
	storage := NewStorage()
	account := &Account{
		Address:       address,
		Node:          node,
		Storage:       storage,
	}
	return account, nil
}

func NewContractAccount(callerAddr CommonAddress, chainId, nonce int64) (*Account, error) {
	address := GetByteCodeAddress(callerAddr, nonce)
	storage := NewStorage()
	account := &Account{
		Address: address,
		Node: &Node{ChainId: chainId},
		Storage: storage,
	}
	return account, nil
}
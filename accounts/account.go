package accounts

import (
	"math/big"
	"BlockChainTest/mycrypto"
	"errors"
	"fmt"
)

var (
	SeedSize = 64
	SeedMark = []byte("Drep Coin Seed")
	KeyBitSize = 256 >> 3
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
		seed, err := genSeed()
		if err != nil {
			return nil
		}
		h := hmAC(seed, SeedMark)
		fmt.Println("h: ", h)
		fmt.Println("len h: ", len(h))
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
	Balance    *big.Int
	Nonce      int64
	ByteCode   ByteCode
	CodeHash   Hash
}

func NewStorage(byteCode ByteCode) *Storage {
	storage := &Storage{}
	storage.Balance = new(big.Int)
	storage.Nonce = 0
	storage.ByteCode = byteCode
	if byteCode != nil {
		storage.CodeHash = GetByteCodeHash(byteCode)
	}
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
	err := store(node)
	if err != nil {
		return nil, err
	}
	address := node.Address()
	storage := NewStorage(nil)
	account := &Account{
		Address:       address,
		Node:          node,
		Storage:       storage,
	}
	return account, nil
}

func NewContractAccount(callerAddr CommonAddress, chainId, nonce int64, byteCode ByteCode) (*Account, error) {
	address := GetByteCodeAddress(callerAddr, nonce)
	storage := NewStorage(byteCode)
	account := &Account{
		Address: address,
		Node: &Node{ChainId: chainId},
		Storage: storage,
	}
	return account, nil
}
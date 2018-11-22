package accounts

import (
	"math/big"
	"BlockChainTest/mycrypto"
)

var (
	SeedSize = 64
	SeedMark = []byte("Drep Coin Seed")
	KeyBitSize = 256
)

type Node struct {
	ChainId ChainID
	PrvKey  *mycrypto.PrivateKey
	Address CommonAddress
}

func NewNode(prv []byte, chainId ChainID) *Node {
	prvKey := genPrvKey(prv)
	address := PubKey2Address(prvKey.PubKey)
	return &Node{
		ChainId: chainId,
		PrvKey: prvKey,
		Address: address,
	}
}

type Storage struct {
	Balance    *big.Int
	Nonce      int64
	IsContract bool
	ByteCode   ByteCode
	CodeHash   Hash
}

func NewStorage(byteCode ByteCode) *Storage {
	storage := &Storage{}
	storage.Balance = new(big.Int)
	storage.Nonce = 0
	storage.ByteCode = byteCode
	if byteCode != nil {
		storage.IsContract = true
		storage.CodeHash = GetCodeHash(byteCode)
	}
	return storage
}

type Account interface {
	Address() CommonAddress
	GetNode() *Node
	GetStorage() *Storage
}

func NewAccount(m *MainAccount, chainId ChainID, byteCode ByteCode) (Account, error) {
	 isMain := chainId == MainChainID
	 if isMain {
	 	return NewMainAccount(byteCode)
	 } else {
	 	return NewSubAccount(m, chainId, byteCode)
	 }
}

type MainAccount struct {
	Node        *Node
	Storage     *Storage
	ChainCode   []byte
	SubAccounts map[ChainID] *SubAccount
}

func NewMainAccount(byteCode ByteCode) (*MainAccount, error) {
	seed, err := genSeed()
	if err != nil {
		return nil, err
	}
	h := hmAC(seed, SeedMark)
	account := &MainAccount{
		Node: NewNode(h[:KeyBitSize], MainChainID),
		Storage: NewStorage(byteCode),
		ChainCode: h[KeyBitSize:],
		SubAccounts: make(map[ChainID] *SubAccount),
	}
	err = store(account.Node)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (m *MainAccount) Address() CommonAddress {
	return m.Node.Address
}

func (m *MainAccount) GetNode() *Node {
	return m.Node
}

func (m *MainAccount) GetStorage() *Storage {
	return m.Storage
}

type SubAccount struct {
	Node    *Node
	Storage *Storage
}

func NewSubAccount(m *MainAccount, chainId ChainID, byteCode ByteCode) (*SubAccount, error) {
	chainCode := new(big.Int).SetBytes(m.ChainCode)
	id := new(big.Int).SetInt64(int64(chainId))
	msg := new(big.Int).Xor(chainCode, id).Bytes()
	h := hmAC(msg, m.Node.PrvKey.Prv)
	account := &SubAccount{
		Node: NewNode(h[:KeyBitSize], chainId),
		Storage: NewStorage(byteCode),
	}
	m.SubAccounts[chainId] = account
	err := store(account.Node)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *SubAccount) Address() CommonAddress {
	return s.Node.Address
}

func (s *SubAccount) GetNode() *Node {
	return s.Node
}

func (s *SubAccount) GetStorage() *Storage {
	return s.Storage
}
package account

import (
	"log"
	"math/big"
	"BlockChainTest/mycrypto"
	"crypto/rand"
	"crypto/hmac"
	"crypto/sha512"
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

func NewStorage(code ByteCode) *Storage {
	storage := &Storage{}
	storage.Balance = new(big.Int)
	storage.Nonce = 0
	storage.ByteCode = code
	if code != nil {
		storage.IsContract = true
		storage.CodeHash = Bytes2Hash(code)
	}
	return storage
}

type Account interface {
	Address() CommonAddress
}

type MainAccount struct {
	Node        *Node
	Storage     *Storage
	ChainCode   []byte
	SubAccounts map[ChainID] *SubAccount
}

func NewMainAccount() (*MainAccount, error) {
	seed, err := genSeed()
	if err != nil {
		return nil, err
	}
	h := hmAC(seed, SeedMark)
	account := &MainAccount{
		Node: NewNode(h[:KeyBitSize], MainChainID),
		Storage: NewStorage(nil),
		ChainCode: h[KeyBitSize:],
		SubAccounts: make(map[ChainID] *SubAccount),
	}
	return account, store(account.Node)
}

func (m *MainAccount) Address() CommonAddress {
	return m.Node.Address
}

func (m *MainAccount) NewSubAccount(chainId ChainID) (*SubAccount, error) {
	code := new(big.Int).SetBytes(m.ChainCode)
	id := new(big.Int).SetInt64(int64(chainId))
	msg := new(big.Int).Xor(code, id).Bytes()
	h := hmAC(msg, m.Node.PrvKey.Prv)
	account := &SubAccount{
		Node: NewNode(h[:KeyBitSize], chainId),
		Storage: NewStorage(nil),
	}
	m.SubAccounts[chainId] = account
	return account, store(account.Node)
}

type SubAccount struct {
	Node    *Node
	Storage *Storage
}

func (s *SubAccount) Address() CommonAddress {
	return s.Node.Address
}

func hmAC(message, key []byte) []byte {
	h := hmac.New(sha512.New, key)
	h.Write(message)
	return h.Sum(nil)
}

func genSeed() ([]byte, error) {
	seed := make([]byte, SeedSize)
	_, err := rand.Read(seed)
	if err != nil {
		log.Println("Error in genSeed().")
	}
	return seed, err
}

func genPrvKey(prv []byte) *mycrypto.PrivateKey {
	cur := mycrypto.GetCurve()
	pubKey := cur.ScalarBaseMultiply(prv)
	prvKey := &mycrypto.PrivateKey{Prv: prv, PubKey: pubKey}
	return prvKey
}
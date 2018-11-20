package account

import (
	"log"
	"math/big"
	"BlockChainTest/mycrypto"
	"BlockChainTest/bean"
	"crypto/rand"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
)

var (
	SeedSize = 64
	SeedMark = []byte("Drep Coin Seed")
	KeyBitSize = 256
)

type ChainID int64

type Node struct {
	PrvKey  *mycrypto.PrivateKey
	Address bean.CommonAddress
}

func NewNode(prv []byte) *Node {
	prvKey := GenPrvKey(prv)
	address := bean.PubKey2Address(prvKey.PubKey)
	return &Node{
		PrvKey: prvKey,
		Address: address,
	}
}

func store(node *Node) error {
	key := &Key{
		PrivateKey: hex.EncodeToString(node.PrvKey.Prv),
		Address: node.Address.Hex(),
	}
	b, err := json.Marshal(key)
	if err != nil {
		return err
	}
	return GenKeystore(key.Address, b)
}

func load(addr string) (*Node, error) {
	jsonBytes, err := LoadKeystore(addr)
	if err != nil {
		return nil, err
	}
	key := &Key{}
	err = json.Unmarshal(jsonBytes, key)
	if err != nil {
		return nil, err
	}
	prv, err := hex.DecodeString(key.PrivateKey)
	if err != nil {
		return nil, err
	}
	node := &Node{
		PrvKey: GenPrvKey(prv),
		Address: bean.Hex2Address(key.Address),
	}
	return node, nil
}

type MainAccount struct {
	Node        *Node
	ChainCode   []byte
	SubAccounts map[ChainID] *SubAccount
}

type SubAccount struct {
	Node *Node
}

func NewMainAccount() (*MainAccount, error) {
	seed, err := GenSeed()
	if err != nil {
		return nil, err
	}
	h := HMAC(seed, SeedMark)
	account := &MainAccount{
		Node: NewNode(h[:KeyBitSize]),
		ChainCode: h[KeyBitSize:],
		SubAccounts: make(map[ChainID] *SubAccount),
	}
	return account, store(account.Node)
}

func (m *MainAccount) NewSubAccount(chainId ChainID) (*SubAccount, error) {
	code := new(big.Int).SetBytes(m.ChainCode)
	id := new(big.Int).SetInt64(int64(chainId))
	msg := new(big.Int).Xor(code, id).Bytes()
	h := HMAC(msg, m.Node.PrvKey.Prv)
	account := &SubAccount{
		Node: NewNode(h[:KeyBitSize]),
	}
	m.SubAccounts[chainId] = account
	return account, store(account.Node)
}

func HMAC(message, key []byte) []byte {
	h := hmac.New(sha512.New, key)
	h.Write(message)
	return h.Sum(nil)
}

func GenSeed() ([]byte, error) {
	seed := make([]byte, SeedSize)
	_, err := rand.Read(seed)
	if err != nil {
		log.Println("Error in GenSeed().")
	}
	return seed, err
}

func GenPrvKey(prv []byte) *mycrypto.PrivateKey {
	cur := mycrypto.GetCurve()
	pubKey := cur.ScalarBaseMultiply(prv)
	prvKey := &mycrypto.PrivateKey{Prv: prv, PubKey: pubKey}
	return prvKey
}
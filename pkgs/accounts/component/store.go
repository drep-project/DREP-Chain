package component

import (
	"encoding/json"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/aes"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"errors"
)

type KeyStore interface {
	// Loads and decrypts the key from disk.
	GetKey(addr *crypto.CommonAddress, auth string) (*chainTypes.Node, error)
	// Writes and encrypts the key.
	StoreKey(k *chainTypes.Node, auth string) error
	// Writes and encrypts the key.
	ExportKey(auth string) ([]*chainTypes.Node, error)
	// Joins filename with the key directory unless it is already absolute.
	JoinPath(filename string) string
}

type CryptedNode struct {
	CryptoPrivateKey []byte                `json:"cryptoPrivateKey"`
	PrivateKey       *secp256k1.PrivateKey `json:"-"`
	ChainId          app.ChainIdType    `json:"chainId"`
	ChainCode        []byte                `json:"chainCode"`

	Key []byte `json:"-"`
	Iv  []byte `json:"iv"`
}

func (cryptedNode *CryptedNode) EnCrypt() {
	cryptedNode.CryptoPrivateKey = aes.AesCBCEncrypt(cryptedNode.PrivateKey.Serialize(), cryptedNode.Key, cryptedNode.Iv)
}

func (cryptedNode *CryptedNode) DeCrypt() *chainTypes.Node {
	privKeyBytes := aes.AesCBCDecrypt(cryptedNode.CryptoPrivateKey, cryptedNode.Key, cryptedNode.Iv)
	privkey, pubkey := secp256k1.PrivKeyFromBytes(privKeyBytes)
	address := crypto.PubKey2Address(pubkey)
	return &chainTypes.Node{
		Address:    &address,
		PrivateKey: privkey,
		ChainId:    cryptedNode.ChainId,
		ChainCode:  cryptedNode.ChainCode,
	}
}

// bytesToCryptoNode cocnvert given bytes and password to a node
func bytesToCryptoNode(data []byte, auth string) (node *chainTypes.Node, errRef error) {
	defer func() {
		if err := recover(); err != nil {
			errRef = errors.New("decryption failed")
		}
	}()
	cryptoNode := new(CryptedNode)
	if err := json.Unmarshal(data, cryptoNode); err != nil {
		return nil, err
	}
	cryptoNode.Key = []byte(auth)
	node = cryptoNode.DeCrypt()
	return
}


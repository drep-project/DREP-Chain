package component

import (
	"encoding/json"
	accountTypes "github.com/drep-project/drep-chain/accounts/types"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/crypto/aes"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/pkg/errors"
)

type KeyStore interface {
	// Loads and decrypts the key from disk.
	GetKey(addr *crypto.CommonAddress, auth string) (*accountTypes.Node, error)
	// Writes and encrypts the key.
	StoreKey(k *accountTypes.Node, auth string) error
	// Writes and encrypts the key.
	ExportKey(auth string) ([]*accountTypes.Node, error)
	// Joins filename with the key directory unless it is already absolute.
	JoinPath(filename string) string
}

type CryptedNode struct {
	CryptoPrivateKey []byte                `json:"cryptoPrivateKey"`
	PrivateKey       *secp256k1.PrivateKey `json:"-"`
	ChainId          common.ChainIdType    `json:"chainId"`
	ChainCode        []byte                `json:"chainCode"`

	Key []byte `json:"-"`
	Iv  []byte `json:"iv"`
}

func (cryptedNode *CryptedNode) EnCrypt() {
	cryptedNode.CryptoPrivateKey = aes.AesCBCEncrypt(cryptedNode.PrivateKey.Serialize(), cryptedNode.Key, cryptedNode.Iv)
}

func (cryptedNode *CryptedNode) DeCrypt() *accountTypes.Node {
	privKeyBytes := aes.AesCBCDecrypt(cryptedNode.CryptoPrivateKey, cryptedNode.Key, cryptedNode.Iv)
	privkey, pubkey := secp256k1.PrivKeyFromBytes(privKeyBytes)
	address := crypto.PubKey2Address(pubkey)
	return &accountTypes.Node{
		Address:    &address,
		PrivateKey: privkey,
		ChainId:    cryptedNode.ChainId,
		ChainCode:  cryptedNode.ChainCode,
	}
}

// bytesToCryptoNode cocnvert given bytes and password to a node
func bytesToCryptoNode(data []byte, auth string) (node *accountTypes.Node, errRef error) {
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


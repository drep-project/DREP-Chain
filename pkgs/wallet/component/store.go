package component

import (
	"encoding/json"
	"github.com/drep-project/drep-chain/crypto/aes"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	walletTypes "github.com/drep-project/drep-chain/pkgs/wallet/types"
	"errors"
)

type KeyStore interface {
	// Loads and decrypts the key from disk.
	GetKey(pubkey *secp256k1.PublicKey, auth string) (*walletTypes.Key, error)
	// Writes and encrypts the key.
	StoreKey(k *walletTypes.Key, auth string) error
	// Writes and encrypts the key.
	ExportKey(auth string) ([]*walletTypes.Key, error)
	// Joins filename with the key directory unless it is already absolute.
	JoinPath(filename string) string
}

type CryptedNode struct {
	CryptoPrivateKey []byte                	`json:"cryptoPrivateKey"`
	PrivateKey       *secp256k1.PrivateKey 	`json:"-"`
	Key []byte `json:"-"`
	Iv  []byte `json:"iv"`
}

func (cryptedNode *CryptedNode) EnCrypt() {
	cryptedNode.CryptoPrivateKey = aes.AesCBCEncrypt(cryptedNode.PrivateKey.Serialize(), cryptedNode.Key, cryptedNode.Iv)
}

func (cryptedNode *CryptedNode) DeCrypt() *walletTypes.Key {
	privKeyBytes := aes.AesCBCDecrypt(cryptedNode.CryptoPrivateKey, cryptedNode.Key, cryptedNode.Iv)
	privkey, pubkey := secp256k1.PrivKeyFromBytes(privKeyBytes)
	return &walletTypes.Key{
		Pubkey: pubkey,
		PrivKey: privkey,
	}
}

// bytesToCryptoNode cocnvert given bytes and password to a node
func bytesToCryptoNode(data []byte, auth string) (key *walletTypes.Key, errRef error) {
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
	key = cryptoNode.DeCrypt()
	return
}


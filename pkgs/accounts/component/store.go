package component

import (
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
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

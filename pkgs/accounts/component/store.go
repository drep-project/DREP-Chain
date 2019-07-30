package component

import (
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/types"
)

type KeyStore interface {
	// Loads and decrypts the key from disk.
	GetKey(addr *crypto.CommonAddress, auth string) (*types.Node, error)
	// Writes and encrypts the key.
	StoreKey(k *types.Node, auth string) error
	// Writes and encrypts the key.
	ExportKey(auth string) ([]*types.Node, error)
	// Joins filename with the key directory unless it is already absolute.
	JoinPath(filename string) string
}

package component

import (
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"sync"
	"errors"
	"path/filepath"
	walletTypes "github.com/drep-project/drep-chain/pkgs/wallet/types"
)

// accountCache This is used for buffering real storage and upper applications to speed up reading.
// TODO If the write speed becomes a bottleneck, write caching can be added
type CacheStore struct {
	store       KeyStore //  This points to a de facto storage.
	keyStoreDir string
	keys        []*walletTypes.Key
	rlock       sync.RWMutex
}

// NewCacheStore receive an path and password as argument
// path refer to  the file that contain all key
// password used to decrypto content in key file
func NewCacheStore(keyStoreDir string, password string) (*CacheStore, error) {
	store, err := NewFileStore(keyStoreDir, password)
	if err != nil {
		return nil, err
	}
	cacheStore := &CacheStore{
		keyStoreDir: keyStoreDir,
		store:       store,
	}

	persistedNodes, err := cacheStore.store.ExportKey(password)
	if err != nil {
		return nil, err
	}
	cacheStore.keys = persistedNodes
	return cacheStore, nil
}

// GetKey Get the private key by address and password
// Notice if you wallet is locked ,private key cant be found
func (cacheStore *CacheStore) GetKey(pubkey *secp256k1.PublicKey, auth string) (*walletTypes.Key, error) {
	cacheStore.rlock.RLock()
	defer cacheStore.rlock.RUnlock()

	for _, key := range cacheStore.keys {
		if key.Pubkey.IsEqual(pubkey) {
			return key, nil
		}
	}
	return nil, errors.New("key not found")
}

// ExportKey export all key in cache by password
func (cacheStore *CacheStore) ExportKey(auth string) ([]*walletTypes.Key, error) {
	return cacheStore.keys, nil
}

// StoreKey store key local storage medium
func (cacheStore *CacheStore) StoreKey(k *walletTypes.Key, auth string) error {
	cacheStore.rlock.Lock()
	defer cacheStore.rlock.Unlock()

	err := cacheStore.store.StoreKey(k, auth)
	if err != nil {
		return errors.New("save key failed" + err.Error())
	}
	cacheStore.keys = append(cacheStore.keys, k)
	return nil
}

func (cacheStore *CacheStore) ReloadKeys(auth string) error {
	cacheStore.rlock.Lock()
	defer cacheStore.rlock.Unlock()

	for _, key := range cacheStore.keys {
		if key.PrivKey == nil {
			key, err := cacheStore.store.GetKey(key.Pubkey, auth)
			if err != nil {
				return err
			} else {
				key.PrivKey = key.PrivKey
			}
		}
	}
	return nil
}

func (cacheStore *CacheStore) ClearKeys() {
	cacheStore.rlock.Lock()
	defer cacheStore.rlock.Unlock()

	for _, node := range cacheStore.keys {
		node.PrivKey = nil
	}
}

// JoinPath refer to local file
func (cacheStore *CacheStore) JoinPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(cacheStore.keyStoreDir, filename)
}
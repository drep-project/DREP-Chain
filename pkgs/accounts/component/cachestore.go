package component

import (
	"sync"

	"github.com/drep-project/drep-chain/types"
	"github.com/drep-project/drep-chain/crypto"
)

// accountCache This is used for buffering real storage and upper applications to speed up reading.
// TODO If the write speed becomes a bottleneck, write caching can be added
type CacheStore struct {
	store KeyStore //  This points to a de facto storage.
	nodes []*types.Node
	rlock sync.RWMutex
}

// NewCacheStore receive an path and password as argument
// path refer to  the file that contain all key
// password used to decrypto content in key file
func NewCacheStore(keyStore KeyStore, password string) (*CacheStore, error) {
	cacheStore := &CacheStore{
		store: keyStore,
	}
	persistedNodes, err := cacheStore.store.ExportKey(password)
	if err != nil {
		return nil, err
	}
	cacheStore.nodes = persistedNodes
	return cacheStore, nil
}

// GetKey Get the private key by address and password
// Notice if you wallet is locked ,private key cant be found
func (cacheStore *CacheStore) GetKey(addr *crypto.CommonAddress, auth string) (*types.Node, error) {
	cacheStore.rlock.RLock()
	defer cacheStore.rlock.RUnlock()

	for _, node := range cacheStore.nodes {
		if node.Address.Hex() == addr.Hex() {
			return node, nil
		}
	}
	return nil, ErrKeyNotFound
}

// ExportKey export all key in cache by password
func (cacheStore *CacheStore) ExportKey(auth string) ([]*types.Node, error) {
	return cacheStore.nodes, nil
}

// StoreKey store key local storage medium
func (cacheStore *CacheStore) StoreKey(k *types.Node, auth string) error {
	cacheStore.rlock.Lock()
	defer cacheStore.rlock.Unlock()

	err := cacheStore.store.StoreKey(k, auth)
	if err != nil {
		return ErrSaveKey
	}
	cacheStore.nodes = append(cacheStore.nodes, k)
	return nil
}

func (cacheStore *CacheStore) ReloadKeys(auth string) error {
	cacheStore.rlock.Lock()
	defer cacheStore.rlock.Unlock()

	for _, node := range cacheStore.nodes {
		if node.PrivateKey == nil {
			key, err := cacheStore.store.GetKey(node.Address, auth)
			if err != nil {
				return ErrPassword
			} else {
				node.PrivateKey = key.PrivateKey
			}
		}
	}
	return nil
}

func (cacheStore *CacheStore) ClearKeys() {
	cacheStore.rlock.Lock()
	defer cacheStore.rlock.Unlock()

	for _, node := range cacheStore.nodes {
		node.PrivateKey = nil
	}
}

// JoinPath refer to local file
func (cacheStore *CacheStore) JoinPath(filename string) string {
	return cacheStore.store.JoinPath(filename)
}

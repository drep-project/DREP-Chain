package component

import (
	"sync"

	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/types"
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

	cacheStore.nodes = make([]*types.Node, 0)
	return cacheStore, nil
}

// GetKey Get the private key by address and password
// Notice if you wallet is locked ,private key cant be found
func (cacheStore *CacheStore) GetKey(addr *crypto.CommonAddress) (*types.Node, error) {
	cacheStore.rlock.RLock()
	defer cacheStore.rlock.RUnlock()

	for _, node := range cacheStore.nodes {
		if node.Address.Hex() == addr.Hex() {
			return node, nil
		}
	}

	return nil, ErrKeyNotExistOrUnlock
}

func (cacheStore *CacheStore) ListAddr(auth string) ([]string, error) {
	return cacheStore.store.ExportAddrs(auth)
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

// add private key to buff
func (cacheStore *CacheStore) LoadKeys(addr *crypto.CommonAddress, auth string) error {
	cacheStore.rlock.Lock()
	defer cacheStore.rlock.Unlock()

	node, err := cacheStore.store.GetKey(addr, auth)
	if err != nil {
		return err
	}

	for index, node := range cacheStore.nodes {
		if node.Address.String() == addr.String() {
			cacheStore.nodes[index] = node
			return nil
		}
	}

	cacheStore.nodes = append(cacheStore.nodes, node)
	return nil
}

func (cacheStore *CacheStore) ClearKey(addr *crypto.CommonAddress) error {
	cacheStore.rlock.Lock()
	defer cacheStore.rlock.Unlock()

	for key, node := range cacheStore.nodes {
		if node.Address.String() == addr.String() {
			//Quickly deletes an element from a struct array
			cacheStore.nodes = append(cacheStore.nodes[:key], cacheStore.nodes[key+1:]...)
			//node.PrivateKey = nil
			return nil
		}
	}
	return ErrLocked
}

// JoinPath refer to local file
func (cacheStore *CacheStore) JoinPath(filename string) string {
	return cacheStore.store.JoinPath(filename)
}

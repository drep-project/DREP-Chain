package component

import (
	"encoding/json"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/crypto"
	"sync"
)

type MemoryStore struct {
	keys *sync.Map
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{new(sync.Map)}
}

func (mstore *MemoryStore) GetKey(addr *crypto.CommonAddress, auth string) (*chainTypes.Node, error) {
	node, ok := mstore.keys.Load(addr.String())
	if ok {
		return copy(node.(*chainTypes.Node)), nil
	} else {
		return nil, ErrKeyNotFound
	}

}

func (mstore *MemoryStore) StoreKey(k *chainTypes.Node, auth string) error {
	k = copy(k)
	mstore.keys.Store(k.Address.String(), k)
	return nil
}

func (mstore *MemoryStore) ExportKey(auth string) ([]*chainTypes.Node, error) {
	nodes := []*chainTypes.Node{}
	mstore.keys.Range(func(key, value interface{}) bool {
		nodes = append(nodes, copy(value.(*chainTypes.Node)))
		return true
	})
	return nodes, nil
}

func (mstore *MemoryStore) JoinPath(filename string) string {
	return ""
}

func copy(node *chainTypes.Node) *chainTypes.Node {
	bytes, _ := json.Marshal(node)
	newNode := &chainTypes.Node{}
	json.Unmarshal(bytes, newNode)
	return newNode
}

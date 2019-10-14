package chain

import (
	"encoding/json"
	"fmt"
	"github.com/drep-project/drep-chain/chain/store"
	"io/ioutil"
)

// Process genesis contract
type IGenesisProcess interface {
	Genesis(context *GenesisContext) error
}

// GenesisContext used to deliver config information  through all genesis processor
type GenesisContext struct {
	config map[string]json.RawMessage
	store  store.StoreInterface
}

func NewGenesisContext(gensisPath string, store store.StoreInterface) (*GenesisContext, error) {
	fileContent, err := ioutil.ReadFile(gensisPath)
	if err != nil {
		return nil, err
	}
	result := map[string]json.RawMessage{}
	err = json.Unmarshal(fileContent, &result)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &GenesisContext{result, store}, nil
}

func (g GenesisContext) Config() map[string]json.RawMessage {
	return g.config
}

func (g GenesisContext) Store() store.StoreInterface {
	return g.store
}

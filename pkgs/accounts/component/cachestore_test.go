package component

import (
	"bytes"
	"github.com/aristanetworks/goarista/test"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/types"
	"testing"
)

func Test_CacheStoreAndGetKey(t *testing.T) {
	password := "password"
	store := NewMemoryStore()
	cacheStore, _ := NewCacheStore(store, password)
	count := 10

	type TestData struct {
		Query      *crypto.CommonAddress
		WantResult *types.Node
		WantError  error
	}
	tests := make([]*TestData, count)
	for i := 0; i < count; i++ {
		node := types.NewNode(nil, app.ChainIdType{})
		cacheStore.StoreKey(node, password)
		tests[i] = &TestData{
			Query:      node.Address,
			WantResult: node,
		}
	}
	for i := 0; i < count; i++ {
		node := types.NewNode(nil, app.ChainIdType{})
		tests = append(tests, &TestData{
			Query:      node.Address,
			WantResult: nil,
			WantError:  ErrKeyNotFound,
		})
	}

	for _, tData := range tests {
		node, err := cacheStore.GetKey(tData.Query, password)
		if err != nil {
			if err != tData.WantError {
				t.Error("key missing in store：", tData.Query)
			}
		} else {
			if !test.DeepEqual(node, tData.WantResult) {
				t.Errorf("result mismatch for query %v\ngot %v\nwant %v", tData.Query, node, tData.WantResult)
			}
		}
	}
}

func Test_CacheExportKey(t *testing.T) {

	password := "password"
	store := NewMemoryStore()
	cacheStore, _ := NewCacheStore(store, password)
	count := 10

	type TestData struct {
		Query      *crypto.CommonAddress
		WantResult *types.Node
		WantError  error
	}
	tests := make([]*TestData, count)
	for i := 0; i < count; i++ {
		node := types.NewNode(nil, app.ChainIdType{})
		cacheStore.StoreKey(node, password)
		tests[i] = &TestData{
			Query:      node.Address,
			WantResult: node,
		}
	}
	for i := 0; i < count; i++ {
		node := types.NewNode(nil, app.ChainIdType{})
		tests = append(tests, &TestData{
			Query:      node.Address,
			WantResult: nil,
			WantError:  ErrKeyNotFound,
		})
	}

	keys, _ := cacheStore.ExportKey(password)
	for _, tData := range tests {
		isFind := false
		for _, key := range keys {
			if test.DeepEqual(tData.WantResult, key) {
				isFind = true
				break
			}
		}
		if tData.WantError != nil {
			if isFind {
				t.Errorf("key not save but find")
			}
		} else {
			if !isFind {
				t.Errorf("key not save but not find")
			}
		}
	}

}

func Test_CacheClearAndReloadKeys(t *testing.T) {
	password := "password"
	store := NewMemoryStore()
	cacheStore, _ := NewCacheStore(store, password)
	count := 10

	type TestData struct {
		Addr    *crypto.CommonAddress
		Privkey []byte
	}
	tests := make([]*TestData, count)
	for i := 0; i < count; i++ {
		node := types.NewNode(nil, app.ChainIdType{})
		cacheStore.StoreKey(node, password)
		tests[i] = &TestData{
			Addr:    node.Address,
			Privkey: node.PrivateKey.Serialize(),
		}
	}

	cacheStore.ClearKeys()
	keys, _ := cacheStore.ExportKey(password)
	for _, node := range keys {
		if node.PrivateKey != nil {
			t.Error("find privatekey but expect nil")
		}
	}

	cacheStore.ReloadKeys(password)
	keys, _ = cacheStore.ExportKey(password)
	for _, tData := range tests {
		isFind := false
		var node *types.Node
		for _, key := range keys {
			if test.DeepEqual(tData.Addr, key.Address) {
				isFind = true
				node = key
				break
			}
		}

		if isFind {
			if !bytes.Equal(tData.Privkey, node.PrivateKey.Serialize()) {
				t.Error("reload got wrong privatekey expect ", tData.Privkey, " but got ", node.PrivateKey)
			}
		} else {
			t.Error("reload must get privkey but got nothing")
		}
	}
}

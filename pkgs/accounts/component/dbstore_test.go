package component

import (
	"bytes"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/types"
	"testing"
)

func init() {

}

func Test_DBGetKey(t *testing.T) {
	fileStore := NewDbStore("test_db1")
	defer func() {
		fileStore.Close()
		clear("test_db1")
	}()

	node := types.NewNode(nil, app.ChainIdType{})
	err := fileStore.StoreKey(node, pass)
	if err != nil {
		log.Fatal(err)
	}

	gotNode, err := fileStore.GetKey(node.Address, pass)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(gotNode.PrivateKey.Serialize(), node.PrivateKey.Serialize()) {
		t.Error("get key not match the key just saved", gotNode.PrivateKey, node.PrivateKey)
	}
}

func Test_DBExportKey(t *testing.T) {
	fileStore := NewDbStore("test_db2")
	defer func() {
		fileStore.Close()
		clear("test_db2")
	}()

	count := 10
	nodes := make([]*types.Node, count)
	for i := 0; i < count; i++ {
		node := types.NewNode(nil, app.ChainIdType{})
		err := fileStore.StoreKey(node, pass)
		if err != nil {
			log.Fatal(err)
		}
		nodes[i] = node
	}

	gotNodes, err := fileStore.ExportKey(pass)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotNodes) != count {
		t.Error("some key missing after export", len(gotNodes), "  ", count)
	}

	for _, gNode := range nodes {
		isFind := false
		for _, gotNode := range gotNodes {
			if bytes.Equal(gotNode.PrivateKey.Serialize(), gNode.PrivateKey.Serialize()) {
				isFind = true
				break
			}
		}
		if !isFind {
			t.Error("key not found in ketstore", "missing key", gNode)
		}
	}
}

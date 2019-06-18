package component

import (
	"bytes"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/chain/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var (
	pass = "hesl"
)

func init(){

}

func clear(ketStore string){
	fileInfo, _ := ioutil.ReadDir(ketStore)
	for _, file := range fileInfo {
		path := filepath.Join(ketStore, file.Name())
		os.Remove(path)
	}
	os.Remove(ketStore)
}

func Test_GetKey(t *testing.T) {
	fileStore := NewFileStore("test_filestore")
	defer clear("test_filestore")

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

func Test_ExportKey(t *testing.T) {
	fileStore := NewFileStore("test_filestore2")
	defer clear("test_filestore2")

	count := 10
	nodes:= make([]*types.Node, count)
	for i:=0;i<count;i++{
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
		t.Error("some key missing after export",len(gotNodes), "  ", count )
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

func Test_JoinPath(t *testing.T) {
    keystorePath := "test_filestore3"
	defer clear("test_filestore3")
	testData := map[string]string {
		"AA":keystorePath + "\\AA",
		"EC":keystorePath + "\\EC",
		"FR":keystorePath + "\\FR",
		"BHT":keystorePath + "\\BHT",
		"4TGBA":keystorePath + "\\4TGBA",
	}
	fileStore := NewFileStore(keystorePath)

	for path, expectedPath := range testData {
		joinPath := fileStore.JoinPath(path)
		if joinPath != expectedPath {
			t.Error("join path not equal expected path", joinPath, " != ", expectedPath)
		}
	}
}
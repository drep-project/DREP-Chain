package component

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/types"

	crypto2 "github.com/drep-project/DREP-Chain/crypto"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var (
	pass = "hesl"
)

func init() {

}

func clear(ketStore string) {
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

	node := types.NewNode(nil, 1)
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
	nodes := make([]*types.Node, count)
	for i := 0; i < count; i++ {
		node := types.NewNode(nil, 1)
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

func Test_JoinPath(t *testing.T) {
	keystorePath := "test_filestore3"
	defer clear("test_filestore3")
	testData := map[string]string{
		"AA":    keystorePath + "\\AA",
		"EC":    keystorePath + "\\EC",
		"FR":    keystorePath + "\\FR",
		"BHT":   keystorePath + "\\BHT",
		"4TGBA": keystorePath + "\\4TGBA",
	}
	fileStore := NewFileStore(keystorePath)

	for path, expectedPath := range testData {
		joinPath := fileStore.JoinPath(path)
		if joinPath != expectedPath {
			t.Error("join path not equal expected path", joinPath, " != ", expectedPath)
		}
	}
}

func Test_ParsecryptoNode(t *testing.T) {

	//jsonText := `{
	//	"version":0,
	//	"cipherText":null,
	//	"chainId":0,
	//	"chainCode":"MxpSh50fnPuDrAsan0XQyijHPNWU+gpxzAB5ZB1giF2AJAj5fXhn05a6a3BVG2LOD0BRPy0hy0EoP+L8gcxnNw==",
	//	"cipher":"aes-128-ctr",
	//	"cipherParams":{"iv":null},
	//	"KDFParams":{"n":262144,"r":8,"p":1,"dklen":32,"salt":null},
	//	"mac":null
	//}`

	jsonText := `{"version":0,"data":"fvrnWyXJ5rUKhQFkUIM1A2YGYRI+2Quitk8VWzQaWGQ=","cipherText":"Ge9/PHiHCdbo3K7xmYgAdPTGsuTNMLhLxUwjxK74DdU=","chainId":0,"chainCode":"pYBF+yLXFL5hWMAvjBhg0sOwpxCc+tT5WxXKb6oTGmaN3G3Ex/uH+y6DCuM0a+/4h42YilmwuyzkZ5ODaQpntw==","cipher":"aes-128-ctr","cipherParams":{"iv":"GeR3tPhsaruo4RGsziXq8w=="},"KDFParams":{"n":262144,"r":8,"p":1,"dklen":32,"salt":"Qd0cAxqPDyG89XZkeV2luCA22tbxokhG5syaqTLuA7g="},"mac":"ucd5PAASpt02nXDqrBdIL3QLdV8wrdx3UZ/Yij2EQMU="}`

	cryptoNode := new(CryptedNode)
	err := json.Unmarshal([]byte(jsonText), cryptoNode)
	if err != nil {
		fmt.Println(err)
	}

	privD, errRef := DecryptData(*cryptoNode, "123")
	if errRef != nil {
		fmt.Println(err)
	}
	priv, pub := secp256k1.PrivKeyFromScalar(privD)
	addr := crypto2.PubkeyToAddress(pub)
	bytesPriv, _ := json.Marshal(priv)
	fmt.Println(string(bytesPriv))

	bytespub, _ := json.Marshal(pub)
	fmt.Println(string(bytespub))

	fmt.Println(addr.String())

	fmt.Println(string(cryptoNode.ChainCode))
}

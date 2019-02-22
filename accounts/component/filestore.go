package component

import (
	"encoding/json"
	"fmt"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/log"
	accountTypes "github.com/drep-project/drep-chain/accounts/types"
	"io/ioutil"
	"os"
	"path/filepath"
)

type FileStore struct {
	keysDirPath string
}

func NewFileStore(keyStoreDir string) FileStore {
	if !common.IsDirExists(keyStoreDir) {
		err := os.Mkdir(keyStoreDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	return FileStore{
		keysDirPath: keyStoreDir,
	}
}

// GetKey read key in file
func (fs FileStore) GetKey(addr *crypto.CommonAddress, auth string) (*accountTypes.Node, error) {
	contents, err := ioutil.ReadFile(fs.JoinPath(addr.Hex()))
	if err != nil {
		return nil, err
	}

	node, err := bytesToCryptoNode(contents, auth)
	if err != nil {
		return nil, err
	}

	//ensure ressult after read and decrypto correct
	if node.Address.Hex() != addr.Hex() {
		return nil, fmt.Errorf("key content mismatch: have address %x, want %x", node.Address, addr)
	}
	return node, nil
}

// store the key in file encrypto
func (fs FileStore) StoreKey(key *accountTypes.Node, auth string) error {
	iv, err := common.GenUnique()
	if err != nil {
		return err
	}
	cryptoNode := &CryptedNode{
		PrivateKey: key.PrivateKey,
		ChainId:    key.ChainId,
		ChainCode:  key.ChainCode,
		Key:        []byte(auth),
		Iv:         iv[:16],
	}
	cryptoNode.EnCrypt()
	content, err := json.Marshal(cryptoNode)
	if err != nil {
		return err
	}
	return writeKeyFile(fs.JoinPath(key.Address.Hex()), content)
}

// ExportKey export all key in file by password
func (fs FileStore) ExportKey(auth string) ([]*accountTypes.Node, error) {
	persistedNodes := []*accountTypes.Node{}
	err := common.EachChildFile(fs.keysDirPath, func(path string) (bool, error) {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			log.Error("read key store error ", "Msg", err.Error())
			return false, err
		}

		node, err := bytesToCryptoNode(contents, auth)
		if err != nil {
			return false, err
		}

		if err != nil {
			log.Error("read key store error ", "Msg", err.Error())
			return false, err
		}
		persistedNodes = append(persistedNodes, node)
		return true, nil
	})
	if err != nil {
		return nil, err
	}
	return persistedNodes, nil
}

// JoinPath return keystore directory
func (fs FileStore) JoinPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(fs.keysDirPath, filename)
}

func writeTemporaryKeyFile(file string, content []byte) (string, error) {
	// Create the keystore directory with appropriate permissions
	// in case it is not present yet.
	const dirPerm = 0700
	if err := os.MkdirAll(filepath.Dir(file), dirPerm); err != nil {
		return "", err
	}
	// Atomic write: create a temporary hidden file first
	// then move it into place. TempFile assigns mode 0600.
	f, err := ioutil.TempFile(filepath.Dir(file), "."+filepath.Base(file)+".tmp")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	f.Close()
	return f.Name(), nil
}

func writeKeyFile(file string, content []byte) error {
	name, err := writeTemporaryKeyFile(file, content)
	if err != nil {
		return err
	}
	return os.Rename(name, file)
}
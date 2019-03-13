package component

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/drep-project/dlog"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/common/fileutil"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	walletTypes "github.com/drep-project/drep-chain/pkgs/wallet/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type FileStore struct {
	keysDirPath string
}

func NewFileStore(keyStoreDir string, password string) (*FileStore, error) {
	if !fileutil.IsDirExists(keyStoreDir) {
		fileutil.EnsureDir(keyStoreDir)
		metadataPath := filepath.Join(keyStoreDir, ".metadata")
		fs, err := os.Create(metadataPath)
		if err != nil {
			return nil, err
		}
		checkString := sha3.Hash256([]byte(password))
		_, err = fs.Write(checkString)
		if err != nil {
			return nil, err
		}
	}
	store := &FileStore{
		keysDirPath: keyStoreDir,
	}
	if store.checkPassword(password) {
		return store, nil
	}else{
		return nil, errors.New("error password")
	}
}

// GetKey read key in file
func (fs FileStore) GetKey(pubkey *secp256k1.PublicKey, auth string) (*walletTypes.Key, error) {
	saveKey := getSaveKey(pubkey)
	contents, err := ioutil.ReadFile(fs.JoinPath(saveKey))
	if err != nil {
		return nil, err
	}

	key, err := bytesToCryptoNode(contents, auth)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// store the key in file encrypto
func (fs FileStore) StoreKey(key *walletTypes.Key, auth string) error {
	iv, err := common.GenUnique()
	if err != nil {
		return err
	}
	cryptoNode := &CryptedNode{
		PrivateKey: key.PrivKey,
		Key:        []byte(auth),
		Iv:         iv[:16],
	}
	cryptoNode.EnCrypt()
	content, err := json.Marshal(cryptoNode)
	if err != nil {
		return err
	}
	saveKey := getSaveKey(key.Pubkey)
	return writeKeyFile(fs.JoinPath(saveKey), content)
}

// ExportKey export all key in file by password
func (fs FileStore) ExportKey(auth string) ([]*walletTypes.Key, error) {
	persistedNodes := []*walletTypes.Key{}
	err := fileutil.EachChildFile(fs.keysDirPath, func(path string) (bool, error) {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			dlog.Error("read key store error ", "Msg", err.Error())
			return false, err
		}

		node, err := bytesToCryptoNode(contents, auth)
		if err != nil {
			return false, err
		}

		if err != nil {
			dlog.Error("read key store error ", "Msg", err.Error())
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

// JoinPath return keystore directory
func (fs FileStore) checkPassword(password string) bool {
	checkString := sha3.Hash256([]byte(password))
	metadataPath := filepath.Join(fs.keysDirPath, ".metadata")
	saveCheckBytes, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		return false
	}
	return bytes.Equal(saveCheckBytes, checkString)
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

func getSaveKey(pubkey *secp256k1.PublicKey) string {
	return hex.EncodeToString(sha3.Hash256(pubkey.Serialize()))
}
// ensureSaveKey used to ensure key validate
func ensureSaveKey(saveKey string) string {
	if strings.HasPrefix(saveKey, "0x") {
		return saveKey[2:len(saveKey)]
	}
	return saveKey
}
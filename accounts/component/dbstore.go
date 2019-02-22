package component

import (
	"os"
	"fmt"
	"encoding/json"
	"path/filepath"

	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/log"
	"github.com/syndtr/goleveldb/leveldb"
	accountTypes "github.com/drep-project/drep-chain/accounts/types"
)

// DbStore use leveldb as the storegae
type DbStore struct {
	dbDirPath string
	db        *leveldb.DB
}

func NewDbStore(dbStoreDir string) DbStore {
	if !common.IsDirExists(dbStoreDir) {
		err := os.Mkdir(dbStoreDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	db, err := leveldb.OpenFile("account_db", nil)
	if err != nil {
		panic(err)
	}
	return DbStore{
		dbDirPath: dbStoreDir,
		db:        db,
	}
}

// GetKey read key in db
func (db *DbStore) GetKey(addr crypto.CommonAddress, auth string) (*accountTypes.Node, error) {
	bytes := []byte{0}
	node, err := bytesToCryptoNode(bytes, auth)
	if err != nil {
		return nil, err
	}

	//ensure ressult after read and decrypto correct
	if node.Address.Hex() != addr.Hex() {
		return nil, fmt.Errorf("key content mismatch: have address %x, want %x", node.Address, addr)
	}
	return node, nil
}

// store the key in db after encrypto
func (dbStore *DbStore) StoreKey(key *accountTypes.Node, auth string) error {
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
	addr := crypto.PubKey2Address(key.PrivateKey.PubKey()).Hex()
	return dbStore.db.Put([]byte(addr), content, nil)
}

// ExportKey export all key in db by password
func (dbStore *DbStore) ExportKey(auth string) ([]*accountTypes.Node, error) {
	dbStore.db.NewIterator(nil, nil)
	iter := dbStore.db.NewIterator(nil, nil)
	persistedNodes := []*accountTypes.Node{}
	for iter.Next() {
		value := iter.Value()

		node, err := bytesToCryptoNode(value, auth)
		if err != nil {
			log.Error("read key store error ", "Msg", err)
			continue
		}
		persistedNodes = append(persistedNodes, node)
	}
	return persistedNodes, nil
}

// JoinPath return the db file path
func (dbStore *DbStore) JoinPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(dbStore.dbDirPath, "db") //dbfile fixed datadir
}
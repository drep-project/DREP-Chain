package component

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	chainTypes "github.com/drep-project/drep-chain/types"
	"github.com/drep-project/drep-chain/common/fileutil"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/syndtr/goleveldb/leveldb"
)

// DbStore use leveldb as the storegae
type DbStore struct {
	dbDirPath string
	db        *leveldb.DB
}

func NewDbStore(dbStoreDir string) *DbStore {
	if !fileutil.IsDirExists(dbStoreDir) {
		err := os.Mkdir(dbStoreDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	db, err := leveldb.OpenFile(dbStoreDir, nil)
	if err != nil {
		panic(err)
	}

	return &DbStore{
		dbDirPath: dbStoreDir,
		db:        db,
	}
}

// GetKey read key in db
func (db *DbStore) GetKey(addr *crypto.CommonAddress, auth string) (*chainTypes.Node, error) {
	bytes, err := db.db.Get(addr[:], nil)
	if err != nil {
		return nil, err
	}
	node, err := BytesToCryptoNode(bytes, auth)
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
func (dbStore *DbStore) StoreKey(key *chainTypes.Node, auth string) error {
	cryptoNode := &CryptedNode{
		Version:      0,
		Data:         key.PrivateKey.Serialize(),
		ChainId:      key.ChainId,
		ChainCode:    key.ChainCode,
		Cipher:       "aes-128-ctr",
		CipherParams: CipherParams{},
		KDFParams: ScryptParams{
			N:     StandardScryptN,
			R:     scryptR,
			P:     StandardScryptP,
			Dklen: scryptDKLen,
		},
	}
	cryptoNode.EncryptData([]byte(auth))
	content, err := json.Marshal(cryptoNode)
	if err != nil {
		return err
	}
	addr := crypto.PubKey2Address(key.PrivateKey.PubKey())
	return dbStore.db.Put(addr[:], content, nil)
}

// ExportKey export all key in db by password
func (dbStore *DbStore) ExportKey(auth string) ([]*chainTypes.Node, error) {
	dbStore.db.NewIterator(nil, nil)
	iter := dbStore.db.NewIterator(nil, nil)
	persistedNodes := []*chainTypes.Node{}
	for iter.Next() {
		value := iter.Value()

		node, err := BytesToCryptoNode(value, auth)
		if err != nil {
			log.WithField("Msg", err).Error("read key store error ")
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

func (dbStore *DbStore) Close() {
	dbStore.db.Close()
}

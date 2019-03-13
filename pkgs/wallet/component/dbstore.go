package component

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/drep-project/dlog"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/common/fileutil"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/crypto/sha3"
	walletTypes "github.com/drep-project/drep-chain/pkgs/wallet/types"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"path/filepath"
	"strings"
)

// DbStore use leveldb as the storegae
type DbStore struct {
	dbDirPath string
	db        *leveldb.DB
}

func NewDbStore(dbStoreDir string, password string) (*DbStore, error) {
	if !fileutil.IsDirExists(dbStoreDir) {
		err := os.Mkdir(dbStoreDir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	db, err := leveldb.OpenFile("wallet_db", nil)
	if err != nil {
		return nil, err
	}
	checkString := sha3.Hash256([]byte(password))
	checkDbKey := []byte("checkString")
	saveCheckBytes, err := db.Get(checkDbKey, nil )
	if strings.HasPrefix(err.Error(), "leveldb: not found") {
		err := db.Put(checkDbKey,[]byte(checkString), nil)
		if err != nil {
			return nil, err
		}
	}else{
		if !bytes.Equal(saveCheckBytes, checkString) {
			return nil, errors.New("error password")
		}
	}
	return &DbStore{
		dbDirPath: dbStoreDir,
		db:        db,
	}, nil
}

// GetKey read key in db
func (dbStore *DbStore) GetKey(pubkey *secp256k1.PublicKey, auth string) (*walletTypes.Key, error) {
	saveKey := getSaveKey(pubkey)
	contents, err := dbStore.db.Get([]byte(saveKey), nil)
	if err != nil {
		return nil, err
	}
	key, err := bytesToCryptoNode(contents, auth)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// store the key in db after encrypto
func (dbStore *DbStore) StoreKey(key *walletTypes.Key, auth string) error {
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
	return dbStore.db.Put([]byte(saveKey), content, nil)
}

// ExportKey export all key in db by password
func (dbStore *DbStore) ExportKey(auth string) ([]*walletTypes.Key, error) {
	dbStore.db.NewIterator(nil, nil)
	iter := dbStore.db.NewIterator(nil, nil)
	persistedNodes := []*walletTypes.Key{}
	for iter.Next() {
		value := iter.Value()

		node, err := bytesToCryptoNode(value, auth)
		if err != nil {
			dlog.Error("read key store error ", "Msg", err)
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
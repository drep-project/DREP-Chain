package component

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/drep-project/DREP-Chain/common/fileutil"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/types"
	"github.com/drep-project/binary"
	"github.com/syndtr/goleveldb/leveldb"
)

// DbStore use leveldb as the storegae
type DbStore struct {
	dbDirPath string
	db        *leveldb.DB
	quit      chan struct{}
}

func NewDbStore(dbStoreDir string, quit chan struct{}) *DbStore {
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
		quit:      quit,
	}
}

// GetKey read key in db
func (db *DbStore) GetKey(addr *crypto.CommonAddress, auth string) (*types.Node, error) {
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
func (dbStore *DbStore) StoreKey(key *types.Node, auth string) error {
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
	content, err := binary.Marshal(cryptoNode)
	if err != nil {
		return err
	}
	addr := crypto.PubkeyToAddress(key.PrivateKey.PubKey())
	return dbStore.db.Put(addr[:], content, nil)
}

// ExportKey export all key in db by password
func (dbStore *DbStore) ExportKey(auth string) ([]*types.Node, error) {
	dbStore.db.NewIterator(nil, nil)
	iter := dbStore.db.NewIterator(nil, nil)
	persistedNodes := []*types.Node{}

	for {
		select {
		case <-dbStore.quit:
			return nil, nil
		default:
			if iter.Next() {
				value := iter.Value()

				node, err := BytesToCryptoNode(value, auth)
				if err != nil {
					log.WithField("Msg", err).Error("read key store error ")
					continue
				}
				persistedNodes = append(persistedNodes, node)
			} else {
				break
			}
		}
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

package database

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"encoding/hex"
	"sync"
	"strconv"
	"fmt"
	"BlockChainTest/bean"
	"math/big"
	"BlockChainTest/mycrypto"
	"BlockChainTest/trie"
)

var db *Database
var once sync.Once

type Database struct {
	Name      string
	LevelDB   *leveldb.DB
	Trie      *trie.StateTrie
	StateRoot []byte
}

var databaseName = "local_data"

func NewDatabase() *Database {
	ldb, err := leveldb.OpenFile(databaseName, nil)
	if err != nil {
		panic(err)
	}
	return &Database{
		Name: databaseName,
		LevelDB: ldb,
		Trie: trie.NewStateTrie(),
		StateRoot: nil,
	}
}

func GetDatabase() *Database {
	once.Do(func() {
		if db == nil {
			db = NewDatabase()
		}
	})
	return db
}

func (db *Database) Get(key string) (DBElem, error) {
	k, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}
	b, err := db.LevelDB.Get(k, nil)
	if err != nil {
		return nil, err
	}
	return unmarshal(b)
}

func (db *Database) Put(elem DBElem) (string, []byte, error) {
	key := elem.DBKey()
	k, err := hex.DecodeString(key)
	if err != nil {
		return "", nil, err
	}
	b, err := marshal(elem)
	if err != nil {
		return "", nil, err
	}
	return key, b, db.LevelDB.Put(k, b, nil)
}

func (db *Database) PutInt(key string, value int) {
	if err := db.LevelDB.Put([]byte(key), []byte(strconv.Itoa(value)), nil); err != nil {
		fmt.Println("Error!!!!!!!", err)
	}
}

func (db *Database) GetInt(key string) (int, error) {
	if value, err := db.LevelDB.Get([]byte(key), nil); err == nil {
		if r, err := strconv.Atoi(string(value)); err == nil {
			return r, nil
		} else {
			return 0, err
		}
	} else {
		return 0, err
	}
}

func (db *Database) Delete(key string) error {
	k, err := hex.DecodeString(key)
	if err != nil {
		return err
	}
	return db.LevelDB.Delete(k, nil)
}

func (db *Database) Store(key, value []byte) error {
	return db.LevelDB.Put(key, value, nil)
}

func (db *Database) Load(key []byte) ([]byte, error) {
	return db.LevelDB.Get(key, nil)
}

//func (db *Database) Open() {
//	if ldb, err := leveldb.OpenFile(db.Name, nil); err == nil {
//		db.LevelDB = ldb
//	}
//}

//func (db *Database) Close() {
//	db.LevelDB.Close()
//}

//func (db *Database) Clear()  {
//}

type Iterator struct {
	Itr iterator.Iterator
}

func (db *Database) NewIterator() *Iterator {
	return &Iterator{db.LevelDB.NewIterator(nil, nil)}
}

func (itr *Iterator) Next() bool {
	return itr.Itr.Next()
}

func (itr *Iterator) Key() []byte {
	return itr.Itr.Key()
}

func (itr *Iterator) Value() []byte {
	return itr.Itr.Value()
}

func (itr *Iterator) Elem() (DBElem, error) {
	return unmarshal(itr.Itr.Value())
}

func (itr *Iterator) Release() {
	itr.Itr.Release()
}

func GetBlock(height int64) (*bean.Block, error) {
	db := GetDatabase()
	key := bean.Height2Key(height)
	elem, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	if block, ok := elem.(*bean.Block); ok {
		return block, nil
	} else {
		return nil, ErrWrongBlockKey
	}
}

func GetBlocksFrom(start int64) ([]*bean.Block, error) {
	var (
		currentBlock *bean.Block
		err error
		height = start
		blocks = make([]*bean.Block, 0)
	)
	for err == nil {
		currentBlock, err = GetBlock(start)
		if err == nil {
			blocks = append(blocks, currentBlock)
		}
		height += 1
	}
	return blocks, nil
}

func GetAllBlocks() ([]*bean.Block, error) {
	return GetBlocksFrom(int64(0))
}

func GetHighestBlock() (*bean.Block, error) {
	maxHeight, err := GetMaxHeight()
	if err != nil {
		return nil, err
	}
	return GetBlock(maxHeight)
}

func PutBlock(block *bean.Block) error {
	db := GetDatabase()
	_, _, err := db.Put(block)
	return err
}

func GetMaxHeight() (int64, error) {
	db := GetDatabase()
	key := mycrypto.Hash256([]byte("max_height"))
	value, err := db.Load(key)
	if err != nil {
		return -1, err
	}
	return new(big.Int).SetBytes(value).Int64(), nil
}

func PutMaxHeight(height int64) error {
	db := GetDatabase()
	key := mycrypto.Hash256([]byte("max_height"))
	value := new(big.Int).SetInt64(height).Bytes()
	err := db.Store(key, value)
	if err != nil {
		return err
	}
	db.Trie.Insert(key, value)
	return nil
}

func GetBalance(addr bean.CommonAddress) (*big.Int, error) {
	db := GetDatabase()
	key := mycrypto.Hash256([]byte("balance_" + addr.Hex()))
	value, err := db.Load(key)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(value), nil
}

func PutBalance(addr bean.CommonAddress, balance *big.Int) ([]byte, []byte, error) {
	db := GetDatabase()
	key := mycrypto.Hash256([]byte("balance_" + addr.Hex()))
	value := balance.Bytes()
	err := db.Store(key, value)
	if err != nil {
		return nil, nil, err
	}
	db.Trie.Insert(key, value)
	return key, value, nil
}

func GetNonce(addr bean.CommonAddress) (int64, error) {
	db := GetDatabase()
	key := mycrypto.Hash256([]byte("nonce_" + addr.Hex()))
	value, err := db.Load(key)
	if err != nil {
		return -1, err
	}
	return new(big.Int).SetBytes(value).Int64(), nil
}

func PutNonce(addr bean.CommonAddress, nonce int64) ([]byte, []byte, error) {
	db := GetDatabase()
	key := mycrypto.Hash256([]byte("nonce_" + addr.Hex()))
	value := new(big.Int).SetInt64(nonce).Bytes()
	err := db.Store(key, value)
	if err != nil {
		return nil, nil, err
	}
	db.Trie.Insert(key, value)
	return key, value, nil
}
package database

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"encoding/hex"
	"sync"
	"strconv"
	"fmt"
	"BlockChainTest/mycrypto"
	"math/big"
	"BlockChainTest/bean"
)

var db *Database
var once sync.Once

type Database struct {
	LevelDB *leveldb.DB
	Name    string
}

var databaseName = "local_data"

func NewDatabase() *Database {
	ldb, err := leveldb.OpenFile(databaseName, nil)
	if err != nil {
		panic(err)
	}
	return &Database{ldb, databaseName}
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

func (db *Database) GetBlock(height int64) (*bean.Block, error) {
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

func (db *Database) GetBlocksFrom(start int64) ([]*bean.Block, error) {
	var (
		currentBlock *bean.Block
		err error
		height = start
		blocks = make([]*bean.Block, 0)
	)
	for err == nil {
		currentBlock, err = db.GetBlock(start)
		if err == nil {
			blocks = append(blocks, currentBlock)
		}
		height += 1
	}
	return blocks, nil
}

func (db *Database) GetAllBlocks() ([]*bean.Block, error) {
	return db.GetBlocksFrom(int64(0))
}

func (db *Database) PutBlock(block *bean.Block) error {
	_, _, err := db.Put(block)
	return err
}

func (db *Database) GetMaxHeight() (int64, error) {
	key := mycrypto.Hash256([]byte("max_height"))
	value, err := db.Load(key)
	if err != nil {
		return -1, err
	}
	return new(big.Int).SetBytes(value).Int64(), nil
}

func (db *Database) PutMaxHeight(height int64) error {
	key := mycrypto.Hash256([]byte("max_height"))
	value := new(big.Int).SetInt64(height).Bytes()
	return db.Store(key, value)
}

func (db *Database) GetBalance(addr bean.CommonAddress) (*big.Int, error) {
	key := mycrypto.Hash256([]byte("balance_" + addr.Hex()))
	value, err := db.Load(key)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(value), nil
}

func (db *Database) PutBalance(addr bean.CommonAddress, balance *big.Int) error {
	key := mycrypto.Hash256([]byte("balance_" + addr.Hex()))
	value := balance.Bytes()
	return db.Store(key, value)
}

func (db *Database) GetNonce(addr bean.CommonAddress) (int64, error) {
	key := mycrypto.Hash256([]byte("nonce_" + addr.Hex()))
	value, err := db.Load(key)
	if err != nil {
		return -1, err
	}
	return new(big.Int).SetBytes(value).Int64(), nil
}

func (db *Database) PutNonce(addr bean.CommonAddress, nonce int64) error {
	key := mycrypto.Hash256([]byte("balance_" + addr.Hex()))
	value := new(big.Int).SetInt64(nonce).Bytes()
	return db.Store(key, value)
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

func (itr *Iterator) Key() string {
	return string(itr.Itr.Key())
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
package store

import (
	"github.com/drep-project/drep-chain/common/trie"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/database/dbinterface"
)

type StoreDB struct {
	store  dbinterface.KeyValueStore
	cache  *database.TransactionStore //数据属于storage的缓存，调用flush才会把数据写入到diskDb中
	trie   *trie.SecureTrie           //全局状态树  临时树（临时变量）
	trieDb *trie.Database             //状态树存储到磁盘时，使用到的db
}

func NewStoreDB(store dbinterface.KeyValueStore, cache *database.TransactionStore, trie *trie.SecureTrie, trieDb *trie.Database) *StoreDB {
	return &StoreDB{
		store:  store,
		cache:  cache,
		trie:   trie,
		trieDb: trieDb,
	}
}

func (s *StoreDB) initState() error {
	var err error
	s.trie, err = trie.NewSecure(crypto.Hash{}, s.trieDb)
	return err
}

func (s *StoreDB) Get(key []byte) ([]byte, error) {
	var value []byte
	var err error
	if s.cache != nil {
		value, err = s.cache.Get(key)
	} else {
		value, err = s.trie.TryGet(key)
	}
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *StoreDB) Put(key []byte, value []byte) error {
	if s.cache != nil {
		return s.cache.Put(key, value)
	} else {
		err := s.trie.TryUpdate(key, value)
		if err != nil {
			return err
		}
		_, err = s.trie.Commit(nil)
		return err
	}
}

func (s *StoreDB) Delete(key []byte) error {
	if s.cache != nil {
		err := s.cache.Delete(key)
		if err != nil {
			return err
		}
	}
	s.trie.Delete(key)
	_, err := s.trie.Commit(nil)
	return err
}

func (s *StoreDB) Flush() {
	if s.cache != nil {
		s.cache.Flush()
	}
}

func (s *StoreDB) RevertState(shot *database.SnapShot) {
	s.cache.RevertState(shot)
}

func (s *StoreDB) CopyState() *database.SnapShot {
	return s.cache.CopyState()
}

func (s *StoreDB) getStateRoot() []byte {
	return s.trie.Hash().Bytes()
}

func (s *StoreDB) RecoverTrie(root []byte) bool {
	var err error
	s.trie, err = trie.NewSecure(crypto.Bytes2Hash(root), s.trieDb)
	if err != nil {
		return false
	}
	s.cache = database.NewTransactionStore(s.trie)
	return true
}

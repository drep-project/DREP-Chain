package store

import (
	"github.com/drep-project/DREP-Chain/common/trie"
	"github.com/drep-project/DREP-Chain/crypto"
	"github.com/drep-project/DREP-Chain/database"
	"github.com/drep-project/DREP-Chain/database/dbinterface"
)

type StoreDB struct {
	store  dbinterface.KeyValueStore
	cache  *database.TransactionStore //The data belongs to storage's cache and is written to diskDb by a call flush
	trie   *trie.SecureTrie           //Global state tree temporary tree (temporary variable)
	trieDb *trie.Database             //The db used when the state tree is stored to disk
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
		log.WithField("err", err).Info("new secure")
		return false
	}
	s.cache = database.NewTransactionStore(s.trie)
	return true
}

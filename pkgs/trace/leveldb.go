package trace

import (
	"fmt"
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/common/fileutil"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/drep-project/drep-chain/types"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	TX_PREFIX                 = "TX"
	TX_SEND_HISTORY_PREFIX    = "SEND_TXHISTORY"
	TX_RECEIVE_HISTORY_PREFIX = "RECEIVE_TXHISTORY"
)

// LevelDbStore used to save data to level db, there are 3 kinds of prefix in db.
// "TX" for transaction collection,   							format "TX" + hash
// "SEND_TXHISTORY" for transaction group by sender addr,   	format "SEND_TXHISTORY" + addr + hash
// "RECEIVE_TXHISTORY" for transaction group by receive addr	format "RECEIVE_TXHISTORY" + addr + hash
type LevelDbStore struct {
	path string
	db   *leveldb.DB
}

func NewLevelDbStore(path string) (*LevelDbStore, error) {
	fileutil.EnsureDir(path)
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &LevelDbStore{path, db}, nil
}

func (store *LevelDbStore) ExistRecord(block *types.Block) (bool, error) {
	for _, tx := range block.Data.TxList {
		txHash := tx.TxHash()
		key := store.txKey(txHash)
		_, err := store.db.Get(key, nil)
		if err != nil {
			if err == leveldb.ErrNotFound {
				return false, nil
			} else {
				return false, err
			}
		}
	}
	return true, nil
}

// InsertRecord check block ,if tx exist, save to to history and send history , if to is not nil, save tx receive history
func (store *LevelDbStore) InsertRecord(block *types.Block) {
	for _, tx := range block.Data.TxList {
		rawdata := tx.AsPersistentMessage()
		txHash := tx.TxHash()
		key := store.txKey(txHash)
		err := store.db.Put(key, rawdata, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		from, _ := tx.From()
		sendHistoryKey := store.txSendHistoryKey(from, txHash)
		err = store.db.Put(sendHistoryKey, txHash[:], nil)
		if err != nil {
			return
		}

		to := tx.To()
		if to != nil {
			historyKey := store.txReceiveHistoryKey(to, txHash)
			err = store.db.Put(historyKey, txHash[:], nil)
			if err != nil {
				return
			}
		}
	}
}

func (store *LevelDbStore) DelRecord(block *types.Block) {
	for _, tx := range block.Data.TxList {
		txHash := tx.TxHash()
		key := store.txKey(txHash)
		store.db.Delete(key, nil)
		from, _ := tx.From()
		sendHistoryKey := store.txSendHistoryKey(from, txHash)
		store.db.Delete(sendHistoryKey, nil)

		to := tx.To()
		if to != nil {
			receiveHistoryKey := store.txReceiveHistoryKey(to, txHash)
			store.db.Delete(receiveHistoryKey, nil)
		}
	}
}

func (store *LevelDbStore) GetRawTransaction(txHash *crypto.Hash) ([]byte, error) {
	key := store.txKey(txHash)
	rawData, err := store.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return rawData, nil
}

func (store *LevelDbStore) GetTransaction(txHash *crypto.Hash) (*RpcTransaction, error) {
	rawData, err := store.GetRawTransaction(txHash)
	if err != nil {
		return nil, err
	}
	tx := &types.Transaction{}
	err = binary.Unmarshal(rawData, tx)
	if err != nil {
		return nil, err
	}
	rpcTx := &RpcTransaction{}
	rpcTx.FromTx(tx)
	return rpcTx, nil
}

func (store *LevelDbStore) GetSendTransactionsByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*RpcTransaction {
	txs := []*RpcTransaction{}
	fromIndex := (pageIndex - 1) * pageSize
	endIndex := fromIndex + pageSize
	if endIndex <= 0 {
		return txs
	}
	key := store.txSendHistoryPrefixKey(addr)
	snapShot, err := store.db.GetSnapshot()
	if err != nil {
		return txs
	}

	iter := snapShot.NewIterator(util.BytesPrefix(key), nil)
	count := 0
	defer iter.Release()
	for iter.Next() {
		if count >= fromIndex {
			if count < endIndex {
				hash := &crypto.Hash{}
				err = binary.Unmarshal(iter.Value(), hash)
				if err != nil {
					break
				}
				tx, err := store.GetTransaction(hash)
				if err != nil {
					break
				}
				txs = append(txs, tx)
			} else {
				break
			}
		}
		count++
	}

	return txs
}

func (store *LevelDbStore) GetReceiveTransactionsByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*RpcTransaction {
	txs := []*RpcTransaction{}
	fromIndex := (pageIndex - 1) * pageSize
	endIndex := fromIndex + pageSize
	if endIndex <= 0 {
		return txs
	}
	key := store.txReceiveHistoryPrefixKey(addr)
	snapShot, err := store.db.GetSnapshot()
	if err != nil {
		return txs
	}

	iter := snapShot.NewIterator(util.BytesPrefix(key), nil)
	count := 0
	defer iter.Release()
	for iter.Next() {
		if count >= fromIndex {
			if count < endIndex {
				hash := &crypto.Hash{}
				err = binary.Unmarshal(iter.Value(), hash)
				if err != nil {
					break
				}
				tx, err := store.GetTransaction(hash)
				if err != nil {
					break
				}
				txs = append(txs, tx)
			} else {
				break
			}
		}
		count++
	}

	return txs
}

func (store *LevelDbStore) txKey(hash *crypto.Hash) []byte {
	buf := [34]byte{}
	copy(buf[:2], []byte(TX_PREFIX)[:2])
	copy(buf[2:], hash[:])
	return buf[:]
}

func (store *LevelDbStore) txSendHistoryKey(addr *crypto.CommonAddress, hash *crypto.Hash) []byte {
	buf := [66]byte{}
	copy(buf[:14], []byte(TX_SEND_HISTORY_PREFIX)[:14])
	copy(buf[14:34], addr[:])
	copy(buf[34:], hash[:])
	return buf[:]
}

func (store *LevelDbStore) txSendHistoryPrefixKey(addr *crypto.CommonAddress) []byte {
	buf := [34]byte{}
	copy(buf[:14], []byte(TX_SEND_HISTORY_PREFIX)[:14])
	copy(buf[14:34], addr[:])
	return buf[:]
}

func (store *LevelDbStore) txReceiveHistoryKey(addr *crypto.CommonAddress, hash *crypto.Hash) []byte {
	buf := [69]byte{} //17+20+32 = 37+32 = 69
	copy(buf[:17], []byte(TX_RECEIVE_HISTORY_PREFIX)[:17])
	copy(buf[17:37], addr[:])
	copy(buf[37:], hash[:])
	return buf[:]
}

func (store *LevelDbStore) txReceiveHistoryPrefixKey(addr *crypto.CommonAddress) []byte {
	buf := [37]byte{} //17+32
	copy(buf[:17], []byte(TX_RECEIVE_HISTORY_PREFIX)[:17])
	copy(buf[17:], addr[:])
	return buf[:]
}

func (store *LevelDbStore) Close() {
	store.db.Close()
}

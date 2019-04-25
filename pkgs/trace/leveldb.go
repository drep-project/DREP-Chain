package trace

import (
	"fmt"
	"github.com/drep-project/binary"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common/fileutil"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	TX_PREFIX = "TX"
	TX_SEND_HISTORY_PREFIX    =  "SEND_TXHISTORY"
	TX_RECEIVE_HISTORY_PREFIX =  "RECEIVE_TXHISTORY"
)


type LevelDbStore struct {
	path string
	db        *leveldb.DB
}

func NewLevelDbStore(path string) *LevelDbStore{
	fileutil.EnsureDir(path)
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		panic(err)
	}
	return &LevelDbStore{path,db}
}

func  (store *LevelDbStore) InsertRecord(block *chainTypes.Block)  {
	for _, tx := range block.Data.TxList {
		rawdata := tx.AsPersistentMessage()
		txHash := tx.TxHash()
		key := store.TxKey(txHash)
		err := store.db.Put(key, rawdata, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		from, _ := tx.From()
		sendHistoryKey := store.TxSendHistoryKey(from, txHash)
		err = store.db.Put(sendHistoryKey, txHash[:], nil)
		if err != nil {
			return
		}

		to := tx.To()
		if to != nil {
			historyKey := store.TxReceiveHistoryKey(to, txHash)
			err = store.db.Put(historyKey, txHash[:], nil)
			if err != nil {
				return
			}
		}
	}
}

func (store *LevelDbStore) DelRecord(block *chainTypes.Block)  {
	for _, tx := range block.Data.TxList {
		txHash := tx.TxHash()
		key := store.TxKey(txHash)
		store.db.Delete(key, nil)
		from, _ := tx.From()
		historyKey := store.TxSendHistoryKey(from, txHash)
		store.db.Delete(historyKey, nil)
	}
}

func (store *LevelDbStore) GetRawTransaction(txHash *crypto.Hash) ([]byte, error)  {
	key := store.TxKey(txHash)
	rawData, err := store.db.Get(key,nil)
	if err != nil{
		return nil, err
	}
	return rawData, nil
}

func (store *LevelDbStore) GetTransaction(txHash *crypto.Hash) (*chainTypes.RpcTransaction, error)  {
	rawData, err := store.GetRawTransaction(txHash)
	if err != nil{
		return nil, err
	}
	tx := &chainTypes.Transaction{}
	err = binary.Unmarshal(rawData, tx)
	if err != nil{
		return nil, err
	}
	rpcTx := &chainTypes.RpcTransaction{}
	rpcTx.FromTx(tx)
	return rpcTx, nil
}

func (store *LevelDbStore) GetSendTransactionsByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*chainTypes.RpcTransaction  {
	txs := []*chainTypes.RpcTransaction{}
	fromIndex := (pageIndex - 1) * pageSize
	endIndex := fromIndex + pageSize
	if endIndex  <= 0 {
		return txs
	}
	key := store.TxSendHistoryPrefixKey(addr)
	snapShot, err := store.db.GetSnapshot()
	if err != nil{
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
			}else{
				break
			}
		}
		count ++
	}

	return txs
}

func (store *LevelDbStore) GetReceiveTransactionsByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*chainTypes.RpcTransaction  {
	txs := []*chainTypes.RpcTransaction{}
	fromIndex := (pageIndex - 1) * pageSize
	endIndex := fromIndex + pageSize
	if endIndex  <= 0 {
		return txs
	}
	key := store.TxReceiveHistoryPrefixKey(addr)
	snapShot, err := store.db.GetSnapshot()
	if err != nil{
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
			}else{
				break
			}
		}
		count ++
	}

	return txs
}

func (store *LevelDbStore) TxKey(hash *crypto.Hash) []byte {
	buf := [34]byte{}
	copy(buf[:2],[]byte(TX_PREFIX)[:2])
	copy(buf[2:],hash[:])
	return buf[:]
}

func (store *LevelDbStore) TxSendHistoryKey(addr *crypto.CommonAddress, hash *crypto.Hash) []byte {
	buf := [66]byte{}
	copy(buf[:14],[]byte(TX_SEND_HISTORY_PREFIX)[:14])
	copy(buf[14:34],addr[:])
	copy(buf[34:], hash[:])
	return buf[:]
}

func (store *LevelDbStore) TxSendHistoryPrefixKey(addr *crypto.CommonAddress) []byte {
	buf := [34]byte{}
	copy(buf[:14],[]byte(TX_SEND_HISTORY_PREFIX)[:14])
	copy(buf[14:34],addr[:])
	return buf[:]
}

func (store *LevelDbStore) TxReceiveHistoryKey(addr *crypto.CommonAddress, hash *crypto.Hash) []byte {
	buf := [69]byte{}   //17+20+32 = 37+32 = 69
	copy(buf[:17],[]byte(TX_RECEIVE_HISTORY_PREFIX)[:17])
	copy(buf[17:37],addr[:])
	copy(buf[37:], hash[:])
	return buf[:]
}

func (store *LevelDbStore) TxReceiveHistoryPrefixKey(addr *crypto.CommonAddress) []byte {
	buf := [49]byte{}  //17+32
	copy(buf[:17],[]byte(TX_RECEIVE_HISTORY_PREFIX)[:17])
	copy(buf[18:],addr[:])
	return buf[:]
}

func (store *LevelDbStore)  Close(){
	store.db.Close()
}
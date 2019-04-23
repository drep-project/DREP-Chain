package trace

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/app"
	chainService "github.com/drep-project/drep-chain/chain/service"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common/event"
	"github.com/drep-project/drep-chain/common/fileutil"
	"github.com/drep-project/drep-chain/crypto"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"gopkg.in/urfave/cli.v1"
	"path"
)

const (
	TX_PREFIX = "TX"
	TX_HISTORY_PREFIX = "TXHISTORY"
)
var (
	DefaultHistoryConfig = &HistoryConfig{
		Enable: true,
	}
)

// HistoryService use to record tx data for query
// support get transaction by hash
// support get transaction history of address
type TraceService struct {
	Config *HistoryConfig
	ChainService *chainService.ChainService  `service:"chain"`
	eventNewBlockSub event.Subscription
	newBlockChan     chan *chainTypes.Block

	detachBlockSub event.Subscription
	detachBlockChan     chan *chainTypes.Block

	db        *leveldb.DB
}

func (traceService *TraceService) Name() string {
	return "trace"
}

func (traceService *TraceService) Api() []app.API {
	return []app.API{
		app.API{
			Namespace: "trace",
			Version:   "1.0",
			Service: &TraceApi{
				traceService,
			},
			Public: true,
		},
	}
}

func (traceService *TraceService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{}
}

func (traceService *TraceService)  P2pMessages() map[int]interface{} {
	return map[int]interface{}{}
}

func (traceService *TraceService) Init(executeContext *app.ExecuteContext) error {
	traceService.Config = DefaultHistoryConfig
	err := executeContext.UnmashalConfig(traceService.Name(), traceService.Config)
	if err != nil {
		return err
	}
	traceService.newBlockChan = make(chan *chainTypes.Block, 1000)
	traceService.detachBlockChan = make(chan *chainTypes.Block, 1000)
	homeDir := executeContext.CommonConfig.HomeDir
	traceService.Config.HistoryDir = path.Join(homeDir, "trace")
	fileutil.EnsureDir(traceService.Config.HistoryDir)
	db, err := leveldb.OpenFile(traceService.Config.HistoryDir, nil)
	if err != nil {
		panic(err)
	}
	traceService.db = db
	return nil
}

func (traceService *TraceService) Start(executeContext *app.ExecuteContext) error {
	traceService.eventNewBlockSub = traceService.ChainService.NewBlockFeed.Subscribe(traceService.newBlockChan)
	traceService.detachBlockSub = traceService.ChainService.DetachBlockFeed.Subscribe(traceService.detachBlockChan)
	go traceService.Process()
	return nil
}

func  (traceService *TraceService) Process() error {
	for {
		select {
			case block := <- traceService.newBlockChan:
				traceService.InsertRecord(block)
			case block := <- traceService.detachBlockChan:
				traceService.DelRecord(block)
		}
	}
}
func  (traceService *TraceService) InsertRecord(block *chainTypes.Block)  {
	for _, tx := range block.Data.TxList {
		rawdata, err := binary.Marshal(tx)
		if err != nil {
			return
		}
		txHash := tx.TxHash()
		key := TxKey(txHash)
		err = traceService.db.Put(key, rawdata, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		historyKey := TxHistoryKey(tx.From(), txHash)
		err = traceService.db.Put(historyKey, txHash[:], nil)
		if err != nil {
			fmt.Println(err)
			return
		}

	}
}

func (traceService *TraceService) DelRecord(block *chainTypes.Block)  {
	for _, tx := range block.Data.TxList {
		txHash := tx.TxHash()
		key := TxKey(txHash)
		traceService.db.Delete(key, nil)

		historyKey := TxHistoryKey(tx.From(), txHash)
		traceService.db.Delete(historyKey, nil)
	}
}

func (traceService *TraceService) GetRawTransaction(txHash *crypto.Hash) ([]byte, error)  {
	key := TxKey(txHash)
	rawData, err := traceService.db.Get(key,nil)
	if err != nil{
		return nil, err
	}
	return rawData, nil
}

func (traceService *TraceService) GetTransaction(txHash *crypto.Hash) (*chainTypes.Transaction, error)  {
	rawData, err := traceService.GetRawTransaction(txHash)
	if err != nil{
		return nil, err
	}
	tx := &chainTypes.Transaction{}
	err = binary.Unmarshal(rawData, tx)
	if err != nil{
		return nil, err
	}
	return tx, nil
}

func (traceService *TraceService) GetTransactionsByAddr(addr *crypto.CommonAddress, pageIndex, pageSize int) []*chainTypes.Transaction  {
	txs := []*chainTypes.Transaction{}
	fromIndex := (pageIndex - 1) * pageSize
	endIndex := fromIndex + pageSize
	if endIndex  <= 0 {
		return txs
	}
	key := TxHistoryPrefixKey(addr)
	snapShot, err := traceService.db.GetSnapshot()
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
				tx, err := traceService.GetTransaction(hash)
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

func TxKey(hash *crypto.Hash) []byte {
	buf := [34]byte{}
	copy(buf[:2],[]byte(TX_PREFIX)[:2])
	copy(buf[2:],hash[:])
	return buf[:]
}
func TxHistoryKey(addr *crypto.CommonAddress, hash *crypto.Hash) []byte {
	buf := [61]byte{}
	copy(buf[:9],[]byte(TX_HISTORY_PREFIX)[:9])
	copy(buf[9:29],addr[:])
	copy(buf[29:], hash[:])
	return buf[:]
}

func TxHistoryPrefixKey(addr *crypto.CommonAddress) []byte {
	buf := [29]byte{}
	copy(buf[:9],[]byte(TX_HISTORY_PREFIX)[:9])
	copy(buf[9:29],addr[:])
	return buf[:]
}

func (traceService *TraceService) Stop(executeContext *app.ExecuteContext) error{
	traceService.eventNewBlockSub.Unsubscribe()
	traceService.detachBlockSub.Unsubscribe()
	return nil
}

func (traceService *TraceService) Receive(context actor.Context) { }



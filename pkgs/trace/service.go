package trace

import (
	"errors"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/app"
	chainService "github.com/drep-project/drep-chain/chain/service"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common/event"
	"gopkg.in/urfave/cli.v1"
	"path"
)


var (
	DefaultHistoryConfig = &HistoryConfig{
		Enable:  true,
		DbType:  "leveldb",
		Url:     "mongodb://localhost:27017",
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
	store   IStore

	readyToQuit chan struct{}
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

func (traceService *TraceService) P2pMessages() map[int]interface{} {
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
	traceService.readyToQuit = make(chan struct{})
	homeDir := executeContext.CommonConfig.HomeDir
	traceService.Config.HistoryDir = path.Join(homeDir, "trace")

	if traceService.Config.DbType == "leveldb" {
		traceService.store = NewLevelDbStore(traceService.Config.HistoryDir)
	} else if traceService.Config.DbType == "mongo" {
		traceService.store = NewMongogDbStore(traceService.Config.Url)
	}else{
		return errors.New("not support persistence type")
	}
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
			case <-traceService.readyToQuit:
			default:
				select {
				case block := <- traceService.newBlockChan:
					traceService.store.InsertRecord(block)
				case block := <- traceService.detachBlockChan:
					traceService.store.DelRecord(block)
				   default:
				}
		}
	}
}

func (traceService *TraceService) Stop(executeContext *app.ExecuteContext) error{
	traceService.eventNewBlockSub.Unsubscribe()
	traceService.detachBlockSub.Unsubscribe()
	traceService.readyToQuit <- struct{}{}
	traceService.readyToQuit <- struct{}{}
	traceService.store.Close()
	return nil
}

func (traceService *TraceService) Receive(context actor.Context) {

}



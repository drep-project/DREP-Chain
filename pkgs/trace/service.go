package trace

import (
	"path"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/app"
	chainService "github.com/drep-project/drep-chain/chain/service/chainservice"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common/event"
	"gopkg.in/urfave/cli.v1"
)

var (
	DefaultHistoryConfig = &HistoryConfig{
		Enable: false,
		DbType: "leveldb",
		Url:    "mongodb://localhost:27017",
	}

	EnableTraceFlag = cli.BoolFlag{
		Name:  "enableTrace",
		Usage: "is  trace enable flag",
	}
	DefaultDbName = "drep"
)

// HistoryService use to record tx data for query
// support get transaction by hash
// support get transaction history of sender address
// support get transaction history of sender receiver
type TraceService struct {
	Config       *HistoryConfig
	ChainService chainService.ChainServiceInterface `service:"chain"`

	eventNewBlockSub event.Subscription
	newBlockChan     chan *chainTypes.Block

	detachBlockSub  event.Subscription
	detachBlockChan chan *chainTypes.Block

	store       IStore
	readyToQuit chan struct{}
	apis        []app.API
}

func (traceService *TraceService) Name() string {
	return MODULENAME
}

func (traceService *TraceService) Api() []app.API {
	return traceService.apis
}

func (traceService *TraceService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{HistoryDirFlag, EnableTraceFlag}
}

func (traceService *TraceService) P2pMessages() map[int]interface{} {
	return map[int]interface{}{}
}

// Init used to create connection to storage(leveldb and mongo)
func (traceService *TraceService) Init(executeContext *app.ExecuteContext) error {
	traceService.Config = DefaultHistoryConfig
	homeDir := executeContext.CommonConfig.HomeDir
	traceService.Config.HistoryDir = path.Join(homeDir, "trace")
	err := executeContext.UnmashalConfig(traceService.Name(), traceService.Config)
	if err != nil {
		return err
	}
	ctx := executeContext.Cli
	if ctx.GlobalIsSet(EnableTraceFlag.Name) {
		traceService.Config.Enable = ctx.GlobalBool(EnableTraceFlag.Name)
	}
	if ctx.GlobalIsSet(HistoryDirFlag.Name) {
		traceService.Config.HistoryDir = ctx.GlobalString(HistoryDirFlag.Name)
	}

	if !traceService.Config.Enable {
		return nil
	}
	traceService.newBlockChan = make(chan *chainTypes.Block, 1000)
	traceService.detachBlockChan = make(chan *chainTypes.Block, 1000)

	if traceService.Config.DbType == "leveldb" {
		traceService.store, err = NewLevelDbStore(traceService.Config.HistoryDir)
		if err != nil {
			log.WithField("err", err).WithField("path", traceService.Config.HistoryDir).Error("cannot open db file")
		}
	} else if traceService.Config.DbType == "mongo" {
		traceService.store, err = NewMongogDbStore(traceService.Config.Url,DefaultDbName)
		if err != nil {
			log.WithField("err", err).WithField("url", traceService.Config.Url).Error("try connect mongo fail")
		}
	} else {
		return ErrUnSupportDbType
	}
	if err != nil {
		return err
	}
	traceService.apis = []app.API{
		app.API{
			Namespace: MODULENAME,
			Version:   "1.0",
			Service: &TraceApi{
				traceService,
			},
			Public: true,
		},
	}
	traceService.readyToQuit = make(chan struct{})
	return nil
}

func (traceService *TraceService) Start(executeContext *app.ExecuteContext) error {
	if traceService.Config == nil || !traceService.Config.Enable {
		return nil
	}
	traceService.eventNewBlockSub = traceService.ChainService.NewBlockFeed().Subscribe(traceService.newBlockChan)
	traceService.detachBlockSub = traceService.ChainService.DetachBlockFeed().Subscribe(traceService.detachBlockChan)
	go traceService.Process()
	return nil
}

// Process used to resolve two types of signals,
// newBlockChan is the signal that blocks are added to the chain,
// the other is the detachBlockChan that blocks are withdrawn from the chain.
func (traceService *TraceService) Process() error {
	for {
		select {
		case block := <-traceService.newBlockChan:
			traceService.store.InsertRecord(block)
		case block := <-traceService.detachBlockChan:
			traceService.store.DelRecord(block)
		default:
			select {
			case <-traceService.readyToQuit:
				<-traceService.readyToQuit
				goto STOP
			default:
			}
		}
	}
STOP:
	return nil
}

func (traceService *TraceService) Stop(executeContext *app.ExecuteContext) error {
	if traceService.Config == nil || !traceService.Config.Enable {
		return nil
	}
	if traceService.eventNewBlockSub != nil {
		traceService.eventNewBlockSub.Unsubscribe()
	}
	if traceService.detachBlockSub != nil {
		traceService.detachBlockSub.Unsubscribe()
	}
	if traceService.readyToQuit != nil {
		traceService.readyToQuit <- struct{}{} // tell process to stop in deal all blocks in chanel
		traceService.readyToQuit <- struct{}{} // wait for process is ok to stop
		traceService.store.Close()
	}
	return nil
}

func (traceService *TraceService) Receive(context actor.Context) {

}

func (traceService *TraceService) Rebuild(from, end int) error{
	 currentHeight := traceService.ChainService.BestChain().Height()
	 if uint64(from) > currentHeight {
	 	return nil
	 }
	for i:=from; i< end;i++ {
		block, err := traceService.ChainService.GetBlockByHeight(uint64(from))
		if err != nil {
			return ErrBlockNotFound
		}
		exist, err := traceService.store.ExistRecord(block)
		if err != nil {
			return err
		}
		if exist {
			traceService.store.DelRecord(block)
		}
		traceService.store.InsertRecord(block)
	}
	return nil
}

package trace

import (
	"github.com/drep-project/drep-chain/database"
	consensusService "github.com/drep-project/drep-chain/pkgs/consensus/service"
	"path"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/app"
	chainService "github.com/drep-project/drep-chain/chain"
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
	Config           *HistoryConfig
	ChainService     chainService.ChainServiceInterface `service:"chain"`
	ConsensusService *consensusService.ConsensusService `service:"consensus"`
	DatabaseService  *database.DatabaseService          `service:"database"`
	apis             []app.API
	blockAnalysis    *BlockAnalysis
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
	homeDir := executeContext.CommonConfig.HomeDir
	if len(traceService.Config.HistoryDir) == 0 {
		traceService.Config.HistoryDir = path.Join(homeDir, "trace")
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
	traceService.blockAnalysis = NewBlockAnalysis(*traceService.Config, traceService.ConsensusService, traceService.DatabaseService.LevelDb(), traceService.ChainService.GetBlockByHeight)

	traceService.apis = []app.API{
		app.API{
			Namespace: MODULENAME,
			Version:   "1.0",
			Service: &TraceApi{
				traceService.blockAnalysis, traceService,
			},
			Public: true,
		},
	}
	return nil
}

func (traceService *TraceService) Start(executeContext *app.ExecuteContext) error {
	if traceService.Config == nil || !traceService.Config.Enable {
		return nil
	}
	traceService.blockAnalysis.Start(traceService.ChainService.NewBlockFeed(), traceService.ChainService.DetachBlockFeed())
	return nil
}

func (traceService *TraceService) Stop(executeContext *app.ExecuteContext) error {
	if traceService.Config == nil || !traceService.Config.Enable {
		return nil
	}
	traceService.blockAnalysis.Close()
	return nil
}

func (traceService *TraceService) Receive(context actor.Context) {

}

func (traceService *TraceService) DefaultConfig() *HistoryConfig {
	return DefaultHistoryConfig
}

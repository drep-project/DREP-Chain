package log

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/dlog"
	"github.com/drep-project/drep-chain/app"
	"gopkg.in/urfave/cli.v1"
	"path"
)

var (
	DefaultLogConfig = &LogConfig{
		LogLevel: 3,
	}
)

type LogService struct {
	Config *LogConfig
}

func (logService *LogService) Name() string {
	return "log"
}

func (logService *LogService) Api() []app.API {
	return []app.API{
		app.API{
			Namespace: "log",
			Version:   "1.0",
			Service:   &LogApi{},
			Public:    true,
		},
	}
}

func (logService *LogService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{LogDirFlag, LogLevelFlag, VmoduleFlag, BacktraceAtFlag}
}

func (logService *LogService) P2pMessages() map[int]interface{} {
	return map[int]interface{}{}
}

func (logService *LogService) Init(executeContext *app.ExecuteContext) error {
	logService.Config = DefaultLogConfig
	err := executeContext.UnmashalConfig(logService.Name(), logService.Config)
	if err != nil {
		return err
	}
	logService.setLogConfig(executeContext.Cli, executeContext.CommonConfig.HomeDir)
	return dlog.SetUp(logService.Config.DataDir, logService.Config.LogLevel, logService.Config.Vmodule, logService.Config.BacktraceAt)
}

func (logService *LogService) Start(executeContext *app.ExecuteContext) error {
	return nil
}

func (logService *LogService) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

func (logService *LogService) Receive(context actor.Context) {}

// setLogConfig creates an log configuration from the set command line flags,
func (logService *LogService) setLogConfig(ctx *cli.Context, homeDir string) {
	if ctx.GlobalIsSet(LogLevelFlag.Name) {
		logService.Config.LogLevel = ctx.GlobalInt(LogLevelFlag.Name)
	}

	if ctx.GlobalIsSet(VmoduleFlag.Name) {
		logService.Config.Vmodule = ctx.GlobalString(VmoduleFlag.Name)
	}

	if ctx.GlobalIsSet(BacktraceAtFlag.Name) {
		logService.Config.BacktraceAt = ctx.GlobalString(BacktraceAtFlag.Name)
	}

	//logdir
	if ctx.GlobalIsSet(LogDirFlag.Name) {
		logService.Config.DataDir = ctx.GlobalString(LogDirFlag.Name)
	} else {
		logService.Config.DataDir = path.Join(homeDir, "log")
	}
}

package log

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/app"
	"gopkg.in/urfave/cli.v1"
	"path"
)

type LogService struct {
	config *LogConfig
}

func (logService *LogService) Name() string {
	return "log"
}
func (logService *LogService) Api() []app.API {
	return []app.API{}
}
func (logService *LogService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{LogDirFlag, LogLevelFlag, VmoduleFlag, BacktraceAtFlag}
}

func (logService *LogService)  P2pMessages() map[int]interface{} {
	return map[int]interface{}{}
}

func (logService *LogService) Init(executeContext *app.ExecuteContext) error {
	logService.config = &LogConfig{}
	logService.config.LogLevel = 3
	err := executeContext.UnmashalConfig(logService.Name(), logService.config)
	if err != nil {
		return err
	}
	logService.setLogConfig(executeContext.CliContext, executeContext.CommonConfig.HomeDir)
	return SetUp(logService.config)
}

func (logService *LogService) Start(executeContext *app.ExecuteContext) error {
	return nil
}

func (logService *LogService) Stop(executeContext *app.ExecuteContext) error{
	return nil
}

func (logService *LogService) Receive(context actor.Context) { }
// setLogConfig creates an log configuration from the set command line flags,
func (logService *LogService) setLogConfig(ctx *cli.Context, homeDir string) {
	logService.config = &LogConfig{}
	if ctx.GlobalIsSet(LogLevelFlag.Name) {
		logService.config.LogLevel = ctx.GlobalInt(LogLevelFlag.Name)
	} else {
		logService.config.LogLevel = 3
	}

	if ctx.GlobalIsSet(VmoduleFlag.Name) {
		logService.config.Vmodule = ctx.GlobalString(VmoduleFlag.Name)
	}

	if ctx.GlobalIsSet(BacktraceAtFlag.Name) {
		logService.config.BacktraceAt = ctx.GlobalString(BacktraceAtFlag.Name)
	}

	//logdir
	if ctx.GlobalIsSet(LogDirFlag.Name) {
		logService.config.DataDir = ctx.GlobalString(LogDirFlag.Name)
	} else {
		logService.config.DataDir = path.Join(homeDir, "log")
	}
}


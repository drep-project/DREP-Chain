package log

import (
	"errors"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/app"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/urfave/cli.v1"
	"path"
	"strings"
)

const (
	MODULE = "MODULE"
)

var (
	DefaultLogConfig = &LogConfig{
		LogLevel: 3,
	}
)

type LogService struct {
	Config *LogConfig
	apis []app.API
}

func (logService *LogService) Name() string {
	return "log"
}

func (logService *LogService) Api() []app.API {
	return logService.apis
}

func (logService *LogService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{LogDirFlag, LogLevelFlag, VmoduleFlag}
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
	baseLogPath := path.Join(logService.Config.DataDir, "log")

	wirter1 := &lumberjack.Logger{
		Filename:   baseLogPath,
		MaxSize:    1, // megabytes
		MaxBackups: 300,
		MaxAge:     28, //days
		Compress:   false, // disabled by default
	}
	logrus.SetFormatter(&NullFormat{})
	textFormat := &logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors: true,
	}
	logrus.SetLevel(logrus.Level(logService.Config.LogLevel))
	mHook := NewMyHook(wirter1, &logrus.JSONFormatter{}, textFormat)
	if logService.Config.Vmodule != "" {
		args := []interface{}{}
		pairs := strings.Split(logService.Config.Vmodule, ";")
		for _, pair := range pairs {
			k_v := strings.Split(pair, "=")
			if len(k_v) != 2 {
				return errors.New("not correct module format")
			}
			args = append(args, k_v[0])
			args = append(args, k_v[1])
		}
		mHook.SetModulesLevel(args...)
	}
	logrus.AddHook(mHook)

	logService.apis = []app.API{
		app.API{
			Namespace: "log",
			Version:   "1.0",
			Service:   NewLogApi(mHook),
			Public:    true,
		},
	}
    return nil
	//return dlog.SetUp(logService.Config.DataDir, logService.Config.LogLevel, logService.Config.Vmodule, logService.Config.BacktraceAt)
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
	//logdir
	if ctx.GlobalIsSet(LogDirFlag.Name) {
		logService.Config.DataDir = ctx.GlobalString(LogDirFlag.Name)
	} else {
		logService.Config.DataDir = path.Join(homeDir, "log")
	}
}


func NewLogger(moduleName string) *logrus.Entry {
	return logrus.WithField(MODULE, moduleName)
}


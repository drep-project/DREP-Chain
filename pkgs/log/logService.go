package log

import (
	"errors"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/app"
	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/urfave/cli.v1"
	"os"
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
	loggers = make(map[string]*logrus.Entry)
)

//LogService provier log config and control
type LogService struct {
	Config *LogConfig
	apis   []app.API
}

// Name log name
func (logService *LogService) Name() string {
	return "log"
}

// Name log api to control log function
func (logService *LogService) Api() []app.API {
	return logService.apis
}

// CommandFlags export flag to control log while app running
func (logService *LogService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{LogDirFlag, LogLevelFlag, VmoduleFlag}
}

// P2pMessages no p2p msg for log
func (logService *LogService) P2pMessages() map[int]interface{} {
	return map[int]interface{}{}
}

//Init init log format and output
func (logService *LogService) Init(executeContext *app.ExecuteContext) error {
	logService.setLogConfig(executeContext.Cli, executeContext.CommonConfig.HomeDir)
	baseLogPath := path.Join(logService.Config.DataDir, "log")

	wirter1 := &lumberjack.Logger{
		Filename:   baseLogPath,
		MaxSize:    1, // megabytes
		MaxBackups: 300,
		MaxAge:     28,    //days
		Compress:   false, // disabled by default
	}
	logrus.SetFormatter(&NullFormat{})
	textFormat := &prefixed.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     true,
		ForceFormatting: true,
	}
	logrus.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
	logrus.SetLevel(logrus.TraceLevel)
	mHook := NewMyHook(wirter1, &logrus.JSONFormatter{}, textFormat)
	lv, err := parserLevel(logService.Config.LogLevel)
	if err != nil {
		return err
	}
	for key, _ := range loggers {
		mHook.moduleLevel[key] = lv
	}

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

// Start
func (logService *LogService) Start(executeContext *app.ExecuteContext) error {
	return nil
}

// Stop
func (logService *LogService) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

// Receive
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

// EnsureLogger create logger int other file
func EnsureLogger(moduleName string) *logrus.Entry {
	log, ok := loggers[moduleName]
	if !ok {
		log = logrus.WithFields(logrus.Fields{
			"prefix": moduleName,
			MODULE:   moduleName,
		})
		loggers[moduleName] = log
	}
	return log
}

func (logService *LogService) DefaultConfig() *LogConfig {
	return DefaultLogConfig
}

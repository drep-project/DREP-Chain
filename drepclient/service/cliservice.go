package service

import (
	"fmt"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/drepclient/component/console"
	cliTypes "github.com/drep-project/drep-chain/drepclient/types"
	"github.com/drep-project/drep-chain/log"
	rpcComponent "github.com/drep-project/drep-chain/rpc/component"
	"gopkg.in/urfave/cli.v1"
)

var (
	ConfigFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TODO add config description",
	}
)

// CliService provides an interactive command line window
type CliService struct {
	config *cliTypes.Config
	Log *log.LogService `service:"log"`
}

// Name name
func (cliService *CliService) Name() string {
	return "cli"
}

// Api api none
func (cliService *CliService) Api() []app.API {
	return []app.API{}
}

// Flags flags  enable load js and execute before run
func (cliService *CliService) Flags() []cli.Flag {
	return []cli.Flag{cliTypes.JSpathFlag, cliTypes.ExecFlag, cliTypes.PreloadJSFlag}
}

func (cliService *CliService) P2pMessages() map[int]interface{} {
	return map[int]interface{}{}
}

// Init  set console config
func (cliService *CliService) Init(executeContext *app.ExecuteContext) error {
	endpoint := executeContext.CliContext.Args().First()
	if len(endpoint) == 0 {
		return fmt.Errorf("You have to specify an address")
	}
	client, err := rpcComponent.Dial(endpoint)
	if err != nil {
		return fmt.Errorf("Unable to attach to remote drep: %v", err)
	}

	path := executeContext.CommonConfig.HomeDir
	cliService.config = &cliTypes.Config{}
	cliService.config.Config = console.Config{
		HomeDir: path,
		DocRoot: executeContext.CliContext.GlobalString(cliTypes.JSpathFlag.Name),
		Client:  client,
		Preload: cliTypes.MakeConsolePreloads(executeContext.CliContext),
	}
	return nil
}

func (cliService *CliService) Start(executeContext *app.ExecuteContext) error {
	return cliService.remoteConsole(executeContext)
}

func (cliService *CliService) Stop(executeContext *app.ExecuteContext) error {
	console.Stdin.Close()
	return nil
}

// remoteConsole will connect to a remote drep instance, attaching a JavaScript
// console to it.
func (cliService *CliService) remoteConsole(executeContext *app.ExecuteContext) error {
	console, err := console.New(cliService.config.Config)
	if err != nil {
		return fmt.Errorf("Failed to start the JavaScript console: %v", err)
	}
	defer console.Stop(false)

	if script := executeContext.CliContext.GlobalString(cliTypes.ExecFlag.Name); script != "" {
		console.Evaluate(script)
		return nil
	}

	// Otherwise print the welcome screen and enter interactive mode
	console.Welcome()
	console.Interactive()

	return nil
}

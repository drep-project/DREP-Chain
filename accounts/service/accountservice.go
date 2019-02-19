package service

import (
	"encoding/json"
	"github.com/drep-project/drep-chain/common"
	"gopkg.in/urfave/cli.v1"
	path2 "path"

	accountCommponent "github.com/drep-project/drep-chain/accounts/component"
	accountTypes "github.com/drep-project/drep-chain/accounts/types"
	"github.com/drep-project/drep-chain/app"
)

var (
	KeyStoreDirFlag = common.DirectoryFlag{
		Name:  "keystore",
		Usage: "Directory for the keystore (default = inside the homedir)",
	}
)

// CliService provides an interactive command line window
type AccountService struct {
	config *accountTypes.Config
	Wallet *accountCommponent.Wallet
	apis   []app.API
}

// Name name
func (accountService *AccountService) Name() string {
	return "accounts"
}

// Api api none
func (accountService *AccountService) Api() []app.API {
	return accountService.apis
}

// Flags flags  enable load js and execute before run
func (accountService *AccountService) Flags() []cli.Flag {
	return []cli.Flag{KeyStoreDirFlag}
}

func (accountService *AccountService)  P2pMessages() map[int]interface{} {
	return map[int]interface{}{}
}

// Init  set console config
func (accountService *AccountService) Init(executeContext *app.ExecuteContext) error {
	accountService.config = &accountTypes.Config{}
	config := executeContext.GetConfig(accountService.Name())
	if config != nil {
		err := json.Unmarshal(config, accountService.config)
		if err != nil {
			return err
		}
	}

	if executeContext.CliContext.IsSet(KeyStoreDirFlag.Name) {
		accountService.config.KeyStoreDir = executeContext.CliContext.GlobalString(KeyStoreDirFlag.Name)
	}

	if !path2.IsAbs(accountService.config.KeyStoreDir) {
		if accountService.config.KeyStoreDir == "" {
			accountService.config.KeyStoreDir = path2.Join(executeContext.CommonConfig.HomeDir, "KeyStore")
		} else {
			accountService.config.KeyStoreDir = path2.Join(executeContext.CommonConfig.HomeDir, accountService.config.KeyStoreDir)
		}
	}

	var err error
	accountService.Wallet, err = accountCommponent.NewWallet(accountService.config, accountTypes.RootChain)
	if err != nil {
		return err
	}

	accountService.apis = []app.API{
		app.API{
			Namespace: "account",
			Version:   "1.0",
			Service: &AccountApi{
				Wallet: accountService.Wallet,
			},
			Public: true,
		},
	}
	return nil
}

func (accountService *AccountService) Start(executeContext *app.ExecuteContext) error {
	return nil
}

func (accountService *AccountService) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

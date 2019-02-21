package service

import (
	accountCommponent "github.com/drep-project/drep-chain/accounts/component"
	accountTypes "github.com/drep-project/drep-chain/accounts/types"
	"github.com/drep-project/drep-chain/app"
	chainService "github.com/drep-project/drep-chain/chain/service"
	"github.com/drep-project/drep-chain/common"
	"gopkg.in/urfave/cli.v1"
	path2 "path"
)

var (
	KeyStoreDirFlag = common.DirectoryFlag{
		Name:  "keystore",
		Usage: "Directory for the keystore (default = inside the homedir)",
	}

	WalletPasswordFlag = cli.StringFlag{
		Name:  "walletpassword",
		Usage: "keep wallet open",
	}
)

// CliService provides an interactive command line window
type AccountService struct {
	chainService chainService.ChainService  `service:"chain"`
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
func (accountService *AccountService) CommandFlags() ([]cli.Command, []cli.Flag)  {
	return nil, []cli.Flag{KeyStoreDirFlag, WalletPasswordFlag}
}

func (accountService *AccountService)  P2pMessages() map[int]interface{} {
	return map[int]interface{}{}
}

// Init  set console config
func (accountService *AccountService) Init(executeContext *app.ExecuteContext) error {
	accountService.config = &accountTypes.Config{}
	err := executeContext.UnmashalConfig(accountService.Name(), accountService.config)
	if err != nil {
		return err
	}

	if executeContext.CliContext.IsSet(WalletPasswordFlag.Name) {
		accountService.config.WalletPassword = executeContext.CliContext.GlobalString(WalletPasswordFlag.Name)
	}

	if executeContext.CliContext.IsSet(KeyStoreDirFlag.Name) {
		accountService.config.KeyStoreDir = executeContext.CliContext.GlobalString(KeyStoreDirFlag.Name)
	}

	if !path2.IsAbs(accountService.config.KeyStoreDir) {
		if accountService.config.KeyStoreDir == "" {
			accountService.config.KeyStoreDir = path2.Join(executeContext.CommonConfig.HomeDir, "keystore")
		} else {
			accountService.config.KeyStoreDir = path2.Join(executeContext.CommonConfig.HomeDir, accountService.config.KeyStoreDir)
		}
	}

	accountService.Wallet, err = accountCommponent.NewWallet(accountService.config, accountTypes.RootChain)
	if err != nil {
		return err
	}
	if accountService.config.WalletPassword != "" {
		accountService.Wallet.Open(accountService.config.WalletPassword )
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

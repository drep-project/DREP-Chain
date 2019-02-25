package service

import (
	path2 "path"
	"gopkg.in/urfave/cli.v1"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/common"
	accountTypes "github.com/drep-project/drep-chain/accounts/types"
	chainService "github.com/drep-project/drep-chain/chain/service"
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

	EnableWalletFlag = cli.BoolFlag{
		Name:  "enableWallet",
		Usage: "is wallet flag",
	}
)

// CliService provides an interactive command line window
type AccountService struct {
	ChainService *chainService.ChainService  `service:"chain"`
	config *accountTypes.Config
	Wallet *Wallet
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

	if executeContext.Cli.IsSet(WalletPasswordFlag.Name) {
		accountService.config.WalletPassword = executeContext.Cli.GlobalString(WalletPasswordFlag.Name)
	}

	if executeContext.Cli.IsSet(KeyStoreDirFlag.Name) {
		accountService.config.KeyStoreDir = executeContext.Cli.GlobalString(KeyStoreDirFlag.Name)
	}

	if executeContext.Cli.IsSet(EnableWalletFlag.Name) {
		accountService.config.EnableWallet = executeContext.Cli.GlobalBool(EnableWalletFlag.Name)
	}

	if !accountService.config.EnableWallet {
		return nil
	}

	if !path2.IsAbs(accountService.config.KeyStoreDir) {
		if accountService.config.KeyStoreDir == "" {
			accountService.config.KeyStoreDir = path2.Join(executeContext.CommonConfig.HomeDir, "keystore")
		} else {
			accountService.config.KeyStoreDir = path2.Join(executeContext.CommonConfig.HomeDir, accountService.config.KeyStoreDir)
		}
	}

	accountService.Wallet, err = NewWallet(accountService.config, accountTypes.RootChain)
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
				chainService: accountService.ChainService,
			},
			Public: true,
		},
	}
	return nil
}

func (accountService *AccountService) Start(executeContext *app.ExecuteContext) error {
	if !accountService.config.EnableWallet {
		return nil
	}
	return nil
}

func (accountService *AccountService) Stop(executeContext *app.ExecuteContext) error {
	if !accountService.config.EnableWallet {
		return nil
	}
	return nil
}

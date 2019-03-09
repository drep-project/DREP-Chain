package service

import (
	"errors"
	path2 "path"
	"gopkg.in/urfave/cli.v1"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/common/fileutil"
	accountComponent "github.com/drep-project/drep-chain/pkgs/accounts/component"
	accountTypes "github.com/drep-project/drep-chain/pkgs/accounts/types"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
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
	DatabaseService *database.DatabaseService  `service:"database"`
	ChainService *chainService.ChainService  `service:"chain"`
	CommonConfig *app.CommonConfig
	Config       *accountTypes.Config
	Wallet       *Wallet
	apis         []app.API
}

// Name name
func (accountService *AccountService) Name() string {
	return "accounts"
}

// Api api none
func (accountService *AccountService) Api() []app.API {
	return  []app.API{
		app.API{
			Namespace: "account",
			Version:   "1.0",
			Service: &AccountApi{
				Wallet: accountService.Wallet,
				chainService: accountService.ChainService,
				accountService: accountService,
				databaseService: accountService.DatabaseService,
			},
			Public: true,
		},
	}
}

// Flags flags  enable load js and execute before run
func (accountService *AccountService) CommandFlags() ([]cli.Command, []cli.Flag)  {
	return nil, []cli.Flag{KeyStoreDirFlag, WalletPasswordFlag}
}

func (accountService *AccountService)  P2pMessages() map[int]interface{} {
	return map[int]interface{}{}
}

// Init  set console Config
func (accountService *AccountService) Init(executeContext *app.ExecuteContext) error {
	accountService.CommonConfig = executeContext.CommonConfig
	accountService.Config = &accountTypes.Config{
		KeyStoreDir: path2.Join(executeContext.CommonConfig.HomeDir, "keystore"),
	}
	err := executeContext.UnmashalConfig(accountService.Name(), accountService.Config)
	if err != nil {
		return err
	}

	if executeContext.Cli.IsSet(WalletPasswordFlag.Name) {
		accountService.Config.WalletPassword = executeContext.Cli.GlobalString(WalletPasswordFlag.Name)
	}

	if executeContext.Cli.IsSet(KeyStoreDirFlag.Name) {
		accountService.Config.KeyStoreDir = executeContext.Cli.GlobalString(KeyStoreDirFlag.Name)
	}

	accountService.Wallet, err = NewWallet(accountService.Config, accountService.ChainService.Config.ChainId)
	if err != nil {
		return err
	}
	if accountService.Config.WalletPassword != "" {
		accountService.Wallet.Open(accountService.Config.WalletPassword )
	}
	return nil
}

func (accountService *AccountService) Start(executeContext *app.ExecuteContext) error {
	return nil
}

func (accountService *AccountService) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

func (accountService *AccountService) CreateWallet(password string) error {
	if fileutil.IsDirExists(accountService.Config.KeyStoreDir) {
		if !fileutil.IsEmptyDir(accountService.Config.KeyStoreDir) {
			return errors.New("exist keystore")
		}
	}else {
		fileutil.EnsureDir(accountService.Config.KeyStoreDir)
	}

	store := accountComponent.NewFileStore(accountService.Config.KeyStoreDir)
	password = string(sha3.Hash256([]byte(password)))
	newNode := chainTypes.NewNode(nil, accountService.ChainService.Config.ChainId)
	store.StoreKey(newNode, password)
	return nil
}

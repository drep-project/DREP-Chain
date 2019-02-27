package service

import (
	"errors"
	accountComponent "github.com/drep-project/drep-chain/accounts/component"
	accountTypes "github.com/drep-project/drep-chain/accounts/types"
	"github.com/drep-project/drep-chain/app"
	chainService "github.com/drep-project/drep-chain/chain/service"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto/sha3"
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

	EnableWalletFlag = cli.BoolFlag{
		Name:  "enableWallet",
		Usage: "is wallet flag",
	}
)

// CliService provides an interactive command line window
type AccountService struct {
	ChainService *chainService.ChainService  `service:"chain"`
	CommonConfig *app.CommonConfig
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
	return  []app.API{
		app.API{
			Namespace: "account",
			Version:   "1.0",
			Service: &AccountApi{
				Wallet: accountService.Wallet,
				chainService: accountService.ChainService,
				accountService: accountService,
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

// Init  set console config
func (accountService *AccountService) Init(executeContext *app.ExecuteContext) error {
	accountService.CommonConfig = executeContext.CommonConfig
	accountService.config = &accountTypes.Config{
		KeyStoreDir: path2.Join(executeContext.CommonConfig.HomeDir, "keystore"),
	}
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

	accountService.Wallet, err = NewWallet(accountService.config, accountService.ChainService.Config.ChainId)
	if err != nil {
		return err
	}
	if accountService.config.WalletPassword != "" {
		accountService.Wallet.Open(accountService.config.WalletPassword )
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
	if common.IsDirExists(accountService.config.KeyStoreDir) {
		if !common.IsEmptyDir(accountService.config.KeyStoreDir) {
			return errors.New("exist keystore")
		}
	}else {
		common.EnsureDir(accountService.config.KeyStoreDir)
	}

	store := accountComponent.NewFileStore(accountService.config.KeyStoreDir)
	password = string(sha3.Hash256([]byte(password)))
	newNode := accountTypes.NewNode(nil, accountService.CommonConfig.RootChain)
	store.StoreKey(newNode, password)
	return nil
}

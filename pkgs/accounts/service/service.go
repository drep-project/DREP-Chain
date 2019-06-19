package service

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/chain/service/blockmgr"
	chainTypes "github.com/drep-project/drep-chain/chain/types"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/common/fileutil"
	"github.com/drep-project/drep-chain/crypto/sha3"
	"github.com/drep-project/drep-chain/database"
	accountComponent "github.com/drep-project/drep-chain/pkgs/accounts/component"
	accountTypes "github.com/drep-project/drep-chain/pkgs/accounts/types"
	"gopkg.in/urfave/cli.v1"
	"path/filepath"
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

	DefaultConfig = &accountTypes.Config{
		Enable: true,
		Type:"filestore",
		KeyStoreDir:"keystore",
	}
)


// AccountService
type AccountService struct {
	DatabaseService *database.DatabaseService `service:"database"`
	Blockmgr        *blockmgr.BlockMgr        `service:"blockmgr"`
	CommonConfig    *app.CommonConfig
	Config          *accountTypes.Config
	Wallet          *Wallet
	apis            []app.API
}

// Name service name
func (accountService *AccountService) Name() string {
	return MODULENAME
}

// Api api none
func (accountService *AccountService) Api() []app.API {
	return []app.API{
		app.API{
			Namespace: "account",
			Version:   "1.0",
			Service: &AccountApi{
				Wallet:          accountService.Wallet,
				blockmgr:        accountService.Blockmgr,
				accountService:  accountService,
				databaseService: accountService.DatabaseService,
			},
			Public: true,
		},
	}
}

// Flags flags  enable load js and execute before run
func (accountService *AccountService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{KeyStoreDirFlag, WalletPasswordFlag}
}

func (accountService *AccountService) P2pMessages() map[int]interface{} {
	return map[int]interface{}{}
}

// Init  set console Config
func (accountService *AccountService) Init(executeContext *app.ExecuteContext) error {
	accountService.CommonConfig = executeContext.CommonConfig
	accountService.Config = DefaultConfig
	accountService.Config.KeyStoreDir = filepath.Join(executeContext.CommonConfig.HomeDir, "keystore")
	err := executeContext.UnmashalConfig(accountService.Name(), accountService.Config)
	if err != nil {
		return err
	}

	if executeContext.Cli.GlobalIsSet(EnableWalletFlag.Name) {
		accountService.Config.Enable = executeContext.Cli.GlobalBool(EnableWalletFlag.Name)
	}

	if executeContext.Cli.GlobalIsSet(WalletPasswordFlag.Name) {
		accountService.Config.Password = executeContext.Cli.GlobalString(WalletPasswordFlag.Name)
	}

	if executeContext.Cli.GlobalIsSet(KeyStoreDirFlag.Name) {
		accountService.Config.KeyStoreDir = executeContext.Cli.GlobalString(KeyStoreDirFlag.Name)
	}

	if !accountService.Config.Enable {
		return nil
	}

	accountService.Wallet, err = NewWallet(accountService.Config, accountService.Blockmgr.ChainService.ChainID())
	if err != nil {
		return err
	}
	if accountService.Config.Password != "" {
		err = accountService.Wallet.Open(accountService.Config.Password)
		if err != nil {
			return err
		}
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
			return ErrExistKeystore
		}
	} else {
		fileutil.EnsureDir(accountService.Config.KeyStoreDir)
	}

	store := accountComponent.NewFileStore(accountService.Config.KeyStoreDir)
	password = string(sha3.Keccak256([]byte(password)))
	newNode := chainTypes.NewNode(nil, accountService.Blockmgr.ChainService.Config.ChainId)
	store.StoreKey(newNode, password)
	return nil
}

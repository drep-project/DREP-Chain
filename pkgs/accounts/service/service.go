package service

import (
	"github.com/drep-project/DREP-Chain/params"
	"github.com/drep-project/DREP-Chain/pkgs/evm"
	"path/filepath"

	"github.com/drep-project/DREP-Chain/app"
	"github.com/drep-project/DREP-Chain/blockmgr"
	"github.com/drep-project/DREP-Chain/chain"
	"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/common/fileutil"
	"github.com/drep-project/DREP-Chain/crypto/sha3"
	"github.com/drep-project/DREP-Chain/database"
	accountComponent "github.com/drep-project/DREP-Chain/pkgs/accounts/component"
	accountTypes "github.com/drep-project/DREP-Chain/pkgs/accounts/types"
	chainTypes "github.com/drep-project/DREP-Chain/types"
	"gopkg.in/urfave/cli.v1"
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
		Enable:      true,
		Type:        "filestore",
		KeyStoreDir: "keystore",
	}
)

// AccountService
type AccountService struct {
	EvmService         *evm.EvmService             `service:"vm"`
	DatabaseService    *database.DatabaseService   `service:"database"`
	Chain              chain.ChainServiceInterface `service:"chain"`
	PoolQuery          blockmgr.IBlockMgrPool      `service:"blockmgr"`
	MessageBroadCastor blockmgr.ISendMessage       `service:"blockmgr"`
	Config             *accountTypes.Config
	Wallet             *Wallet
	apis               []app.API
	quit               chan struct{}
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
				EvmService:         accountService.EvmService,
				Wallet:             accountService.Wallet,
				messageBroadCastor: accountService.MessageBroadCastor,
				poolQuery:          accountService.PoolQuery,
				accountService:     accountService,
				databaseService:    accountService.DatabaseService,
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
	if len(accountService.Config.KeyStoreDir) == 0 {
		accountService.Config.KeyStoreDir = filepath.Join(executeContext.CommonConfig.HomeDir, "keystore")
	} else {
		if !filepath.IsAbs(accountService.Config.KeyStoreDir) {
			accountService.Config.KeyStoreDir = filepath.Join(executeContext.CommonConfig.HomeDir, accountService.Config.KeyStoreDir)
		}
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

	//if !accountService.Config.Enable {
	//	return nil
	//}

	accountService.quit = make(chan struct{})

	var err error
	accountService.Wallet, err = NewWallet(accountService.Config, accountService.Chain.ChainID())
	if err != nil {
		return err
	}
	//if accountService.Config.Password != "" {
	err = accountService.Wallet.OpenWallet(accountService.Config.Password)
	if err != nil {
		return err
	}
	//}

	return nil
}

func (accountService *AccountService) Start(executeContext *app.ExecuteContext) error {
	if accountService.Config.Enable {
		return nil
	}
	return nil
}

func (accountService *AccountService) Stop(executeContext *app.ExecuteContext) error {
	if accountService.Config == nil || accountService.Config.Enable {
		return nil
	}
	close(accountService.quit)
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
	newNode := chainTypes.NewNode(nil, accountService.Chain.GetConfig().ChainId)
	store.StoreKey(newNode, password)
	return nil
}

func (accountService *AccountService) DefaultConfig(netType params.NetType) *accountTypes.Config {
	return DefaultConfig
}

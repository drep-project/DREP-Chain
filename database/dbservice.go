package database

import (
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/database/dbinterface"
	"github.com/drep-project/drep-chain/database/leveldb"
	"github.com/drep-project/drep-chain/database/memorydb"
	dlog "github.com/drep-project/drep-chain/pkgs/log"

	"gopkg.in/urfave/cli.v1"
	path2 "path"
)

var (
	MODULENAME = "database"

	log = dlog.EnsureLogger(MODULENAME)

	DataDirFlag = common.DirectoryFlag{
		Name:  "datadir",
		Usage: "Directory for the database dir (default = inside the homedir)",
	}
)

type DatabaseService struct {
	Config *DatabaseConfig
	db     dbinterface.KeyValueStore
}

func NewDatabaseService(db dbinterface.KeyValueStore) *DatabaseService {
	ds := &DatabaseService{db: db}
	return ds
}

func (database *DatabaseService) Name() string {
	return "database"
}

func (database *DatabaseService) Api() []app.API {
	return []app.API{}
}

func (database *DatabaseService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{DataDirFlag}
}

func (database *DatabaseService) Init(executeContext *app.ExecuteContext) error {
	path := path2.Join(executeContext.CommonConfig.HomeDir, "data")
	if executeContext.Cli != nil && executeContext.Cli.GlobalIsSet(DataDirFlag.Name) {
		path = executeContext.Cli.GlobalString(DataDirFlag.Name)
	}
	var err error
	database.db, err = leveldb.New(path, 16, 512, "")
	if err != nil {
		return err
	}
	return nil
}

func (database *DatabaseService) Start(executeContext *app.ExecuteContext) error {
	return nil
}

func (database *DatabaseService) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

func (database *DatabaseService) LevelDb() dbinterface.KeyValueStore {
	return database.db
}

func (database *DatabaseService) MemoryDb() dbinterface.KeyValueStore {
	return memorydb.New()
}

type DatabaseConfig struct {
}

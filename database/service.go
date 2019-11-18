package database

import (
	"github.com/drep-project/DREP-Chain/app"
	"github.com/drep-project/DREP-Chain/common"
	"github.com/drep-project/DREP-Chain/database/trie"
	dlog "github.com/drep-project/DREP-Chain/pkgs/log"

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
	db     *Database
}

func NewDatabaseService(db *Database) *DatabaseService {
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
	database.db, err = NewDatabase(path)
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

func (database *DatabaseService) Db() *Database {
	return database.db
}

func (database *DatabaseService) GetTriedDB() *trie.Database {
	return database.db.trieDb
}
func (database *DatabaseService) RecoverTrie(root []byte) bool {
	return database.db.RecoverTrie(root)
}

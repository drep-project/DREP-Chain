package database

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/common"
	"gopkg.in/urfave/cli.v1"
	path2 "path"
)

var (
	DataDirFlag = common.DirectoryFlag{
		Name:  "datadir",
		Usage: "Directory for the database dir (default = inside the homedir)",
	}
)

type DatabaseService struct {
	config *DatabaseConfig
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

func (database *DatabaseService) Receive(context actor.Context) {}

func (database *DatabaseService) P2pMessages() map[int]interface{} {
	return map[int]interface{}{}
}

func (database *DatabaseService) Init(executeContext *app.ExecuteContext) error {
	err := executeContext.UnmashalConfig(database.Name(), database.config)
	if err != nil {
		return err
	}

	path := path2.Join(executeContext.CommonConfig.HomeDir, "data")
	if executeContext.Cli != nil && executeContext.Cli.IsSet(DataDirFlag.Name) {
		path = executeContext.Cli.GlobalString(DataDirFlag.Name)
	}
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
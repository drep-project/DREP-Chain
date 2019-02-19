package database

import (
	"encoding/json"
	"github.com/drep-project/drep-chain/app"
	"gopkg.in/urfave/cli.v1"
)

type DatabaseService struct {
	config *DatabaseConfig
}


func (database *DatabaseService) Name() string {
	return "database"
}

func (database *DatabaseService) Api() []app.API {
	return []app.API{}
}

func (database *DatabaseService) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (database *DatabaseService)  P2pMessages() map[int]interface{} {
	return map[int]interface{}{}
}

func (database *DatabaseService) Init(executeContext *app.ExecuteContext) error {
	phase := executeContext.GetConfig(database.Name())
	database.config = &DatabaseConfig{}
	err := json.Unmarshal(phase, database.config)
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



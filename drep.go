package main

import (
	"fmt"
	accountService "github.com/drep-project/drep-chain/accounts/service"
	"github.com/drep-project/drep-chain/app"
	chainService "github.com/drep-project/drep-chain/chain/service"
	consensusService "github.com/drep-project/drep-chain/consensus/service"
	"github.com/drep-project/drep-chain/database"
	cliService "github.com/drep-project/drep-chain/drepclient/service"
	"github.com/drep-project/drep-chain/log"
	p2pService "github.com/drep-project/drep-chain/network/service"
	rpcService "github.com/drep-project/drep-chain/rpc/service"
)

func main() {
	drepApp := app.NewApp()
	err := drepApp.AddServices(
		database.DatabaseService{},
		rpcService.RpcService{},
		log.LogService{},
		p2pService.P2pService{},

		chainService.ChainService{},
		accountService.AccountService{},
		consensusService.ConsensusService{},
		cliService.CliService{},
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	drepApp.Name = "drep"
	drepApp.Author = "Drep-Project"
	drepApp.Email = ""
	drepApp.Version = "0.1"
	drepApp.HideVersion = true
	drepApp.Copyright = "Copyright 2018 - now The drep Authors"

	if err := drepApp.Run(); err != nil {
		fmt.Println(err)
	}
	return
}

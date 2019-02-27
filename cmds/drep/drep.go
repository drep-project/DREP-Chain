package drep

import (
	"fmt"
	"github.com/drep-project/drep-chain/log"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/database"
	rpcService "github.com/drep-project/drep-chain/rpc/service"
	chainService "github.com/drep-project/drep-chain/chain/service"
	p2pService "github.com/drep-project/drep-chain/network/service"
	cliService "github.com/drep-project/drep-chain/drepclient/service"
	accountService "github.com/drep-project/drep-chain/accounts/service"
	consensusService "github.com/drep-project/drep-chain/consensus/service"
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

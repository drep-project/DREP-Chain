package main

import (
	"fmt"
	"github.com/drep-project/drep-chain/app"
	blockService "github.com/drep-project/drep-chain/chain/service/blockmgr"
	chainService "github.com/drep-project/drep-chain/chain/service/chainservice"
	"github.com/drep-project/drep-chain/database"
	p2pService "github.com/drep-project/drep-chain/network/service"
	accountService "github.com/drep-project/drep-chain/pkgs/accounts/service"
	consensusService "github.com/drep-project/drep-chain/pkgs/consensus/service"
	cliService "github.com/drep-project/drep-chain/pkgs/drepclient/service"
	evmService "github.com/drep-project/drep-chain/pkgs/evm"
	"github.com/drep-project/drep-chain/pkgs/log"
	"github.com/drep-project/drep-chain/pkgs/rpc"
	"github.com/drep-project/drep-chain/pkgs/trace"
	"runtime/debug"
)

func main() {
	debug.SetGCPercent(20)

	drepApp := app.NewApp()
	err := drepApp.AddServices(
		database.DatabaseService{},

		rpc.RpcService{},
		log.LogService{},
		p2pService.P2pService{},
		evmService.EvmService{},
		chainService.ChainService{},
		blockService.BlockMgr{},
		trace.TraceService{},
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

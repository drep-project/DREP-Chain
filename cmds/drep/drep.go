package main

import (
	"fmt"
	"github.com/drep-project/binary"
	"github.com/drep-project/drep-chain/app"
	blockService "github.com/drep-project/drep-chain/blockmgr"
	chainService "github.com/drep-project/drep-chain/chain"
	"github.com/drep-project/drep-chain/database"
	p2pService "github.com/drep-project/drep-chain/network/service"
	accountService "github.com/drep-project/drep-chain/pkgs/accounts/service"
	chainIndexerService "github.com/drep-project/drep-chain/pkgs/chain_indexer"
	consensusService "github.com/drep-project/drep-chain/pkgs/consensus/service"
	cliService "github.com/drep-project/drep-chain/pkgs/drepclient/service"
	evmService "github.com/drep-project/drep-chain/pkgs/evm"
	filterService "github.com/drep-project/drep-chain/pkgs/filter"
	logServer "github.com/drep-project/drep-chain/pkgs/log"
	"github.com/drep-project/drep-chain/pkgs/rpc"
	"github.com/drep-project/drep-chain/pkgs/trace"
	"log"

	"runtime/debug"
)

func main() {
	debug.SetGCPercent(20)

	drepApp := app.NewApp()
	err := drepApp.IncludeServices(
		database.DatabaseService{},
		rpc.RpcService{},
		logServer.LogService{},
		p2pService.P2pService{},
		chainService.ChainService{},
		blockService.BlockMgr{},
		evmService.EvmService{},
		chainIndexerService.ChainIndexerService{},
		filterService.FilterService{},
		accountService.AccountService{},
		consensusService.ConsensusService{},
		trace.TraceService{},
		cliService.CliService{},
	)

	if err != nil {
		log.Print(err)
		return
	}

	drepApp.Option(func() {
		drepApp.Name = "drep"
		drepApp.Author = "Drep-Project"
		drepApp.Email = ""
		drepApp.Version = "0.1"
		drepApp.HideVersion = true
		drepApp.Copyright = "Copyright 2018 - now The drep Authors"

		binary.MAXSLICESIZE = 1024 * 1024
	})

	if err := drepApp.Run(); err != nil {
		fmt.Println(err)
	}
	return
}

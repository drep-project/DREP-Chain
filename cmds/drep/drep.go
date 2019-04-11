package main

import (
	"fmt"
	"github.com/drep-project/drep-chain/app"
	chainService "github.com/drep-project/drep-chain/chain/service"
	"github.com/drep-project/drep-chain/database"
	p2pService "github.com/drep-project/drep-chain/network/service"
	accountService "github.com/drep-project/drep-chain/pkgs/accounts/service"
	consensusService "github.com/drep-project/drep-chain/pkgs/consensus/service"
	cliService "github.com/drep-project/drep-chain/pkgs/drepclient/service"
	evmService "github.com/drep-project/drep-chain/pkgs/evm"
	"github.com/drep-project/drep-chain/pkgs/log"
	"github.com/drep-project/drep-chain/pkgs/rpc"
	"net/http"
	_ "net/http/pprof"
	"runtime/debug"
)

func main() {
	go func() {
		fmt.Println("http://localhost:8080/debug/pprof")
		http.ListenAndServe("0.0.0.0:8080", nil)
	}()

	debug.SetGCPercent(20)

	drepApp := app.NewApp()
	err := drepApp.AddServices(
		database.DatabaseService{},

		rpc.RpcService{},
		log.LogService{},
		p2pService.P2pService{},
		evmService.EvmService{},
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

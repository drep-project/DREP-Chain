package main

import (
	"fmt"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/pkgs/log"
	"github.com/drep-project/drep-chain/database"
	"github.com/drep-project/drep-chain/pkgs/rpc"
	chainService "github.com/drep-project/drep-chain/chain/service"
	p2pService "github.com/drep-project/drep-chain/network/service"
	cliService "github.com/drep-project/drep-chain/pkgs/drepclient/service"
	accountService "github.com/drep-project/drep-chain/pkgs/wallet/service"
	consensusService "github.com/drep-project/drep-chain/pkgs/consensus/service"
	evmService "github.com/drep-project/drep-chain/pkgs/evm"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() {
		http.ListenAndServe("0.0.0.0:8080", nil)
	}()

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

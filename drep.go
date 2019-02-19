package main

import (
	"fmt"
	"github.com/drep-project/drep-chain/log"
	"reflect"

	accountService "github.com/drep-project/drep-chain/accounts/service"
	chainService "github.com/drep-project/drep-chain/chain/service"
	"github.com/drep-project/drep-chain/app"
	cliService "github.com/drep-project/drep-chain/drepclient/service"
	rpcService "github.com/drep-project/drep-chain/rpc/service"
	p2pService "github.com/drep-project/drep-chain/network/service"
)

func main() {
	drepApp := app.NewApp()
	err := drepApp.AddServiceType(
		reflect.TypeOf(log.LogService{}),
		reflect.TypeOf(p2pService.P2pService{}),
		reflect.TypeOf(chainService.ChainService{}),
		reflect.TypeOf(accountService.AccountService{}),
		reflect.TypeOf(rpcService.RpcService{}),
		reflect.TypeOf(cliService.CliService{}),
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

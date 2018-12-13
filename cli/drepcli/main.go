// Copyright 2014 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

// drep is the official command-line client for Ethereum.
package main

import (
	"fmt"
	"math"
	"os"
	godebug "runtime/debug"
	"sort"
	"strconv"

	"BlockChainTest/cli/drepcli/console"
	"BlockChainTest/config"
	"BlockChainTest/log"
	"BlockChainTest/util/flags"
	"github.com/elastic/gosigar"
	"gopkg.in/urfave/cli.v1"
)

var (
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	// The app that holds all commands and flags.
	app = flags.NewApp(gitCommit, "the drep command line interface")
	nCfg *config.NodeConfig
	nodeFlags = []cli.Flag{
		flags.HomeDirFlag,
		flags.DataDirFlag,
		flags.LogDirFlag,
		flags.KeyStoreDirFlag,
		config.ConfigFileFlag,
		flags.LogLevelFlag,
		flags.VmoduleFlag,
		flags.BacktraceAtFlag,
		flags.ConsensusModeFlag,
	}
	rpcFlags = []cli.Flag{
		flags.HTTPEnabledFlag,
		flags.HTTPListenAddrFlag,
		flags.HTTPPortFlag,
		flags.HTTPApiFlag,
		flags.WSEnabledFlag,
		flags.WSListenAddrFlag,
		flags.WSPortFlag,
		flags.WSApiFlag,
		flags.WSAllowedOriginsFlag,
		flags.IPCDisabledFlag,
		flags.IPCPathFlag,
		flags.RESTEnabledFlag,
		flags.RESTListenAddrFlag,
		flags.RESTPortFlag,
	}
)

func init() {
	// Initialize the CLI app and start Drep
	app.Action = drep
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2013-2018 The drep Authors"
	app.Commands = []cli.Command{
		// See consolecmd.go:
		accountCommand,
		consoleCommand,
		attachCommand,
	//	javascriptCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = append(app.Flags, nodeFlags...)
	app.Flags = append(app.Flags, rpcFlags...)
	app.Flags = append(app.Flags, consoleFlags...)

	app.Before = func(ctx *cli.Context) error {
		var err error
		nCfg, err = config.MakeConfig(ctx)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		err = log.SetUp(&nCfg.LogConfig)  //logDir config here
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		// Cap the cache allowance and tune the garbage collector
		var mem gosigar.Mem
		if err := mem.Get(); err == nil {
			allowance := int(mem.Total / 1024 / 1024 / 3)
			if cache := ctx.GlobalInt(flags.CacheFlag.Name); cache > allowance {
				log.Warn("Sanitizing cache to Go's GC limits", "provided", cache, "updated", allowance)
				ctx.GlobalSet(flags.CacheFlag.Name, strconv.Itoa(allowance))
			}
		}
		// Ensure Go's GC ignores the database cache for trigger percentage
		cache := ctx.GlobalInt(flags.CacheFlag.Name)
		gogc := math.Max(20, math.Min(100, 100/(float64(cache)/1024)))

		log.Debug("Sanitizing Go's GC trigger", "percent", int(gogc))
		godebug.SetGCPercent(int(gogc))
		return nil
	}

	app.After = func(ctx *cli.Context) error {
		console.Stdin.Close() // Resets terminal mode.
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// drep is the main entry point into the system if no special subcommand is ran.
// It creates a default node based on the command line arguments and runs it in
// blocking mode, waiting for it to be shut down.
func drep(ctx *cli.Context) error {
	if args := ctx.Args(); len(args) > 0 {
		return fmt.Errorf("invalid command: %q", args[0])
	}
	nCfg, err := config.MakeConfig(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	//start node and attach
	node := NewNode(nCfg)
	//defer node.Stop()
	node.Start()
	node.Wait()
	return nil
}
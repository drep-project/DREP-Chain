// Copyright 2016 The go-ethereum Authors
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

package main

import (
	"fmt"
	"path"
	"errors"
	"strings"

	"BlockChainTest/rpc"
	"BlockChainTest/config"
	"gopkg.in/urfave/cli.v1"
	"BlockChainTest/util/flags"
	"BlockChainTest/cli/drepcli/console"
)

var (
	consoleFlags = []cli.Flag{flags.JSpathFlag, flags.ExecFlag, flags.PreloadJSFlag}

	consoleCommand = cli.Command{
		Action:   flags.MigrateFlags(localConsole),
		Name:     "console",
		Usage:    "Start an interactive JavaScript environment",
		Flags:    append(append(nodeFlags, rpcFlags...), consoleFlags...),
		Category: "CONSOLE COMMANDS",
		Description: `
The Drep console is an interactive shell for the JavaScript runtime environment
which exposes a node admin interface as well as the Ðapp JavaScript API.
See https://github.com/ethereum/go-ethereum/wiki/JavaScript-Console.`,
	}

	attachCommand = cli.Command{
		Action:    flags.MigrateFlags(remoteConsole),
		Name:      "attach",
		Usage:     "Start an interactive JavaScript environment (connect to node)",
		ArgsUsage: "[endpoint]",
		Flags:     append(consoleFlags, flags.DataDirFlag),
		Category:  "CONSOLE COMMANDS",
		Description: `
The Drep console is an interactive shell for the JavaScript runtime environment
which exposes a node admin interface as well as the Ðapp JavaScript API.
See https://github.com/ethereum/go-ethereum/wiki/JavaScript-Console.
This command allows to open a console on a running drep node.`,
	}

	/*
	javascriptCommand = cli.Command{
		Action:    flags.MigrateFlags(ephemeralConsole),
		Name:      "js",
		Usage:     "Execute the specified JavaScript files",
		ArgsUsage: "<jsfile> [jsfile...]",
		Flags:     consoleFlags,
		Category:  "CONSOLE COMMANDS",
		Description: `
The JavaScript VM exposes a node admin interface as well as the Ðapp
JavaScript API. See https://github.com/ethereum/go-ethereum/wiki/JavaScript-Console`,
	}*/
)

// localConsole starts a new drep node, attaching a JavaScript console to it at the
// same time.
func localConsole(ctx *cli.Context) error {
	if !nCfg.RpcConfig.IPCEnabled {
       return errors.New("ipc must be enable in console mode")
	}
	//start node and attach
	node := NewNode(nCfg)
	//defer node.Stop()
	node.Start()
	<-node.StartComplete
	// Attach to the newly started node and start the JavaScript console
	client, err := node.Attach()
	if err != nil {
		flags.Fatalf("Failed to attach to the inproc drep: %v", err)
	}
	config := console.Config{
		HomeDir: nCfg.HomeDir,
		DocRoot: ctx.GlobalString(flags.JSpathFlag.Name),
		Client:  client,
		Preload: flags.MakeConsolePreloads(ctx),
	}

	console, err := console.New(config)
	if err != nil {
		flags.Fatalf("Failed to start the JavaScript console: %v", err)
	}
	defer console.Stop(false)

	// If only a short execution was requested, evaluate and return
	if script := ctx.GlobalString(flags.ExecFlag.Name); script != "" {
		console.Evaluate(script)
		return nil
	}
	// Otherwise print the welcome screen and enter interactive mode
	console.Welcome()
	console.Interactive()

	return nil
}

// remoteConsole will connect to a remote drep instance, attaching a JavaScript
// console to it.
func remoteConsole(ctx *cli.Context) error {
	// Attach to a remotely running drep instance and start the JavaScript console
	endpoint := ctx.Args().First()
	path := config.DefaultDataDir()
	if endpoint == "" {
		if ctx.GlobalIsSet(flags.DataDirFlag.Name) {
			path = ctx.GlobalString(flags.DataDirFlag.Name)
		}
		endpoint = fmt.Sprintf("%s/drep.ipc", path)
	}
	client, err := dialRPC(nCfg, endpoint)
	if err != nil {
		flags.Fatalf("Unable to attach to remote drep: %v", err)
	}
	config := console.Config{
		HomeDir: path,
		DocRoot: ctx.GlobalString(flags.JSpathFlag.Name),
		Client:  client,
		Preload: flags.MakeConsolePreloads(ctx),
	}

	console, err := console.New(config)
	if err != nil {
		flags.Fatalf("Failed to start the JavaScript console: %v", err)
	}
	defer console.Stop(false)

	if script := ctx.GlobalString(flags.ExecFlag.Name); script != "" {
		console.Evaluate(script)
		return nil
	}

	// Otherwise print the welcome screen and enter interactive mode
	console.Welcome()
	console.Interactive()

	return nil
}

// dialRPC returns a RPC client which connects to the given endpoint.
// The check for empty endpoint implements the defaulting logic
// for "drep attach" and "drep monitor" with no argument.
func dialRPC(cfg *config.NodeConfig, endpoint string) (*rpc.Client, error) {
	if endpoint == "" {
		endpoint = path.Join(cfg.HomeDir, config.DefaultIPCEndpoint(config.ClientIdentifier))  
	} else if strings.HasPrefix(endpoint, "rpc:") || strings.HasPrefix(endpoint, "ipc:") {
		// Backwards compatibility with drep < 1.5 which required
		// these prefixes.
		endpoint = endpoint[4:]
	}
	return rpc.Dial(endpoint)
}

/*
// ephemeralConsole starts a new drep node, attaches an ephemeral JavaScript
// console to it, executes each of the files specified as arguments and tears
// everything down.
func ephemeralConsole(ctx *cli.Context) error {
	// Create and start the node based on the CLI flags
	node := makeFullNode(ctx)
	startNode(ctx, node)
	defer node.Stop()

	// Attach to the newly started node and start the JavaScript console
	client, err := node.Attach()
	if err != nil {
		flags.Fatalf("Failed to attach to the inproc drep: %v", err)
	}
	config := console.Config{
		HomeDir: flags.MakeDataDir(ctx),
		DocRoot: ctx.GlobalString(flags.JSpathFlag.Name),
		Client:  client,
		Preload: flags.MakeConsolePreloads(ctx),
	}

	console, err := console.New(config)
	if err != nil {
		flags.Fatalf("Failed to start the JavaScript console: %v", err)
	}
	defer console.Stop(false)

	// Evaluate each of the specified JavaScript files
	for _, file := range ctx.Args() {
		if err = console.Execute(file); err != nil {
			flags.Fatalf("Failed to execute %s: %v", file, err)
		}
	}
	// Wait for pending callbacks, but stop for Ctrl-C.
	abort := make(chan os.Signal, 1)
	signal.Notify(abort, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-abort
		os.Exit(0)
	}()
	console.Stop(true)

	return nil
}
*/
package main

import (
	"strings"
	"fmt"
	"path"
	"gopkg.in/urfave/cli.v1"
	"BlockChainTest/rpc"
	"BlockChainTest/log"
	"BlockChainTest/cli/drepcli/utils"
)

var (
	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TODO add config description",
	}
)

type nodeConfig struct {
	DataDir string
	ConsensusMode string
	RpcConfig rpc.RpcConfig
	LogConfig log.Config
}


func makeConfig(ctx *cli.Context) ( *nodeConfig) {
	// Load defaults.
	cfg := &nodeConfig{}
	//data dir setting
	setDataDir(ctx, cfg)
	setConsensus(ctx, cfg)
	// TODO Load config file here.
	if file := ctx.GlobalString(configFileFlag.Name); file != "" {
		log.Info("specific file ","PATH", file)
		/*
			if err := loadConfig(file, &cfg); err != nil {
				utils.Fatalf("%v", err)
			}
		*/
	}
	
	// log
	setLogConfig(ctx,cfg)

	//TODO
	//SetP2PConfig(ctx, &cfg.P2P)  

	//rpc Config
	setRpc(ctx, cfg)


	return cfg
}


/*
func SetP2PConfig(ctx *cli.Context, cfg *p2p.Config) {
	
}
*/
// setLogConfig creates an log configuration from the set command line flags,
func setLogConfig(ctx *cli.Context, cfg *nodeConfig) {
	if ctx.GlobalIsSet(utils.LogLevelFlag.Name) {
		cfg.LogConfig.LogLevel = ctx.GlobalInt(utils.LogLevelFlag.Name)
	}else{
		cfg.LogConfig.LogLevel = 3
	}

	if ctx.GlobalIsSet(utils.VmoduleFlag.Name) {
		cfg.LogConfig.Vmodule = ctx.GlobalString(utils.VmoduleFlag.Name)
	}

	if ctx.GlobalIsSet(utils.BacktraceAtFlag.Name) {
		cfg.LogConfig.BacktraceAt = ctx.GlobalString(utils.BacktraceAtFlag.Name)
	}

	cfg.LogConfig.DataDir = path.Join(cfg.DataDir, "log")
}


// setRpc creates an rpc configuration from the set command line flags,
func setRpc(ctx *cli.Context, cfg *nodeConfig) {
	setIPC(ctx, cfg)
	setHTTP(ctx, cfg)
	setWS(ctx, cfg)
	setRest(ctx, cfg)
}


// setIPC creates an IPC path configuration from the set command line flags,
// returning an empty string if IPC was explicitly disabled, or the set path.
func setIPC(ctx *cli.Context, cfg *nodeConfig) {
	cfg.RpcConfig.IPCEnabled = true
	if ctx.GlobalBool(utils.IPCDisabledFlag.Name) {
		cfg.RpcConfig.IPCEnabled = false
		return 
	}
	
	checkExclusive(ctx, utils.IPCDisabledFlag, utils.IPCPathFlag)
	if ctx.GlobalIsSet(utils.IPCPathFlag.Name) {
		cfg.RpcConfig.IPCPath = ctx.GlobalString(utils.IPCPathFlag.Name)
	}else{
		cfg.RpcConfig.IPCPath = DefaultIPCEndpoint(clientIdentifier)
	}
}

// setHTTP creates the HTTP RPC listener interface string from the set
// command line flags, returning empty if the HTTP endpoint is disabled.
func setHTTP(ctx *cli.Context, cfg *nodeConfig) {
	cfg.RpcConfig.HTTPEnabled = true
	if !ctx.GlobalBool(utils.HTTPEnabledFlag.Name) {
		cfg.RpcConfig.HTTPEnabled = false
		return
	} 

	if ctx.GlobalIsSet(utils.HTTPListenAddrFlag.Name) {
		cfg.RpcConfig.HTTPHost = ctx.GlobalString(utils.HTTPListenAddrFlag.Name)
	} else {
		cfg.RpcConfig.HTTPHost = rpc.DefaultHTTPHost
	}

	if ctx.GlobalIsSet(utils.HTTPPortFlag.Name) {
		cfg.RpcConfig.HTTPPort = ctx.GlobalInt(utils.HTTPPortFlag.Name)
	}else{
		cfg.RpcConfig.HTTPPort = rpc.DefaultHTTPPort
	}

	if ctx.GlobalIsSet(utils.HTTPCORSDomainFlag.Name) {
		cfg.RpcConfig.HTTPCors = splitAndTrim(ctx.GlobalString(utils.HTTPCORSDomainFlag.Name))
	}

	if ctx.GlobalIsSet(utils.HTTPApiFlag.Name) {
		cfg.RpcConfig.HTTPModules = splitAndTrim(ctx.GlobalString(utils.HTTPApiFlag.Name))
	}

	if ctx.GlobalIsSet(utils.HTTPVirtualHostsFlag.Name) {
		cfg.RpcConfig.HTTPVirtualHosts = splitAndTrim(ctx.GlobalString(utils.HTTPVirtualHostsFlag.Name))
	} else {
		cfg.RpcConfig.HTTPVirtualHosts = []string{"localhost"}
	}
}

// setHTTP creates the HTTP RPC listener interface string from the set
// command line flags, returning empty if the HTTP endpoint is disabled.
func setRest(ctx *cli.Context, cfg *nodeConfig) {
	cfg.RpcConfig.RESTEnabled = true
	if !ctx.GlobalBool(utils.RESTEnabledFlag.Name) {
		cfg.RpcConfig.RESTEnabled = false
		return
	} 

	if ctx.GlobalIsSet(utils.RESTListenAddrFlag.Name) {
		cfg.RpcConfig.RESTHost = ctx.GlobalString(utils.RESTListenAddrFlag.Name)
	} else {
		cfg.RpcConfig.RESTHost = rpc.DefaultRestHost
	}

	if ctx.GlobalIsSet(utils.RESTPortFlag.Name) {
		cfg.RpcConfig.RESTPort = ctx.GlobalInt(utils.RESTPortFlag.Name)
	}else{
		cfg.RpcConfig.RESTPort = rpc.DefaultRestPort
	}
}

// setWS creates the WebSocket RPC listener interface string from the set
// command line flags, returning empty if the HTTP endpoint is disabled.
func setWS(ctx *cli.Context, cfg *nodeConfig) {

	cfg.RpcConfig.WSEnabled = true
	if !ctx.GlobalBool(utils.WSEnabledFlag.Name) {
		cfg.RpcConfig.WSEnabled = false
		return
	} 

	if ctx.GlobalIsSet(utils.WSListenAddrFlag.Name) {
		cfg.RpcConfig.WSHost = ctx.GlobalString(utils.WSListenAddrFlag.Name)
	} else{
		cfg.RpcConfig.WSHost =  rpc.DefaultWSHost
	}

	if ctx.GlobalIsSet(utils.WSPortFlag.Name) {
		cfg.RpcConfig.WSPort = ctx.GlobalInt(utils.WSPortFlag.Name)
	}else{
		cfg.RpcConfig.WSPort = rpc.DefaultWSPort
	}

	if ctx.GlobalIsSet(utils.WSAllowedOriginsFlag.Name) {
		cfg.RpcConfig.WSOrigins = splitAndTrim(ctx.GlobalString(utils.WSAllowedOriginsFlag.Name))
	}

	if ctx.GlobalIsSet(utils.WSApiFlag.Name) {
		cfg.RpcConfig.WSModules = splitAndTrim(ctx.GlobalString(utils.WSApiFlag.Name))
	}
}

func setConsensus(ctx *cli.Context, cfg *nodeConfig) {
	if ctx.GlobalIsSet(utils.ConsensusModeFlag.Name) {
		cfg.ConsensusMode = ctx.GlobalString(utils.ConsensusModeFlag.Name)
	} else{
		cfg.ConsensusMode = "bft"
	}
}

func setDataDir(ctx *cli.Context, cfg *nodeConfig) {
	if ctx.GlobalIsSet(utils.DataDirFlag.Name) {
		cfg.DataDir = ctx.GlobalString(utils.DataDirFlag.Name)
	} else{
		cfg.DataDir = DefaultDataDir()
	}
}
// checkExclusive verifies that only a single instance of the provided flags was
// set by the user. Each flag might optionally be followed by a string type to
// specialize it further.
func checkExclusive(ctx *cli.Context, args ...interface{}) {
	set := make([]string, 0, 1)
	for i := 0; i < len(args); i++ {
		// Make sure the next argument is a flag and skip if not set
		flag, ok := args[i].(cli.Flag)
		if !ok {
			panic(fmt.Sprintf("invalid argument, not cli.Flag type: %T", args[i]))
		}
		// Check if next arg extends current and expand its name if so
		name := flag.GetName()

		if i+1 < len(args) {
			switch option := args[i+1].(type) {
			case string:
				// Extended flag check, make sure value set doesn't conflict with passed in option
				if ctx.GlobalString(flag.GetName()) == option {
					name += "=" + option
					set = append(set, "--"+name)
				}
				// shift arguments and continue
				i++
				continue

			case cli.Flag:
			default:
				panic(fmt.Sprintf("invalid argument, not cli.Flag or string extension: %T", args[i+1]))
			}
		}
		// Mark the flag if it's set
		if ctx.GlobalIsSet(flag.GetName()) {
			set = append(set, "--"+name)
		}
	}
	if len(set) > 1 {
		utils.Fatalf("Flags %v can't be used at the same time", strings.Join(set, ", "))
	}
}

// splitAndTrim splits input separated by a comma
// and trims excessive white space from the substrings.
func splitAndTrim(input string) []string {
	result := strings.Split(input, ",")
	for i, r := range result {
		result[i] = strings.TrimSpace(r)
	}
	return result
}

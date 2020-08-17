package rpc

import (
	"fmt"
	"github.com/drep-project/DREP-Chain/params"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/DREP-Chain/app"

	"github.com/drep-project/rpc"
	"gopkg.in/urfave/cli.v1"
)

const (
	ClientIdentifier = "drep"
)

type RpcService struct {
	RpcAPIs       []app.API // List of APIs currently provided by the node
	RestApi       rpc.RestDescription
	inprocHandler *rpc.Server // In-process RPC request handler to process the API requests

	IpcEndpoint string       // IPC endpoint to listen at (empty = IPC disabled)
	IpcListener net.Listener // IPC RPC listener socket to serve API requests
	IpcHandler  *rpc.Server  // IPC RPC request handler to process the API requests

	HttpEndpoint  string       // HTTP endpoint (interface + port) to listen at (empty = HTTP disabled)
	HttpWhitelist []string     // HTTP RPC modules to allow through this endpoint
	HttpListener  net.Listener // HTTP RPC listener socket to server API requests
	HttpHandler   *rpc.Server  // HTTP RPC request handler to process the API requests

	WsEndpoint string       // Websocket endpoint (interface + port) to listen at (empty = websocket disabled)
	WsListener net.Listener // Websocket RPC listener socket to server API requests
	WsHandler  *rpc.Server  // Websocket RPC request handler to process the API requests

	RestEndpoint   string              // Websocket endpoint (interface + port) to listen at (empty = websocket disabled)
	RestController *rpc.RestController // Websocket RPC listener socket to server API requests

	lock   sync.RWMutex
	Config *rpc.RpcConfig
}

func (rpcService *RpcService) Name() string {
	return MODULENAME
}

func (rpcService *RpcService) Api() []app.API {
	return nil
}

func (rpcService *RpcService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{
		HTTPEnabledFlag, HTTPListenAddrFlag, HTTPPortFlag, HTTPCORSDomainFlag,
		HTTPVirtualHostsFlag, HTTPApiFlag, IPCDisabledFlag, IPCPathFlag, WSEnabledFlag,
		WSListenAddrFlag, WSPortFlag, WSApiFlag, WSAllowedOriginsFlag, RESTEnabledFlag,
		RESTListenAddrFlag, RESTPortFlag,
	}
}

func (rpcService *RpcService) P2pMessages() map[int]interface{} {
	return map[int]interface{}{}
}

func (rpcService *RpcService) Init(executeContext *app.ExecuteContext) error {
	rpcService.setRpcLog(executeContext.Cli, executeContext.CommonConfig.HomeDir)
	rpcService.IpcEndpoint = rpcService.Config.IPCEndpoint()
	rpcService.HttpEndpoint = rpcService.Config.HTTPEndpoint()
	rpcService.WsEndpoint = rpcService.Config.WSEndpoint()
	rpcService.RestEndpoint = rpcService.Config.RestEndpoint()
	return nil
}

func (rpcService *RpcService) Start(executeContext *app.ExecuteContext) error {
	// All API endpoints started successfully
	rpcService.RpcAPIs = executeContext.GetApis() //api may delay
	// Start the various API endpoints, terminating all in case of errors
	if err := rpcService.StartInProc(rpcService.RpcAPIs); err != nil {
		return err
	}

	if err := rpcService.StartIPC(rpcService.RpcAPIs); err != nil {
		rpcService.StopInProc()
		return err
	}

	if err := rpcService.StartHTTP(rpcService.HttpEndpoint, rpcService.RpcAPIs, rpcService.Config.HTTPModules, rpcService.Config.HTTPCors, rpcService.Config.HTTPVirtualHosts, rpcService.Config.HTTPTimeouts); err != nil {
		rpcService.StopIPC()
		rpcService.StopInProc()
		return err
	}

	if err := rpcService.StartWS(rpcService.WsEndpoint, rpcService.RpcAPIs, rpcService.Config.WSModules, rpcService.Config.WSOrigins, rpcService.Config.WSExposeAll); err != nil {
		rpcService.StopHTTP()
		rpcService.StopIPC()
		rpcService.StopInProc()
		return err
	}

	/*
		if err := rpcService.StartRest(rpcService.RestEndpoint,rpcService.RestApi); err != nil {
			rpcService.StopREST()
			return err
		}
	*/
	return nil
}

func (rpcService *RpcService) Stop(executeContext *app.ExecuteContext) error {
	rpcService.lock.Lock()
	defer rpcService.lock.Unlock()
	// Terminate the API, services and the p2p server.
	rpcService.StopWS()
	rpcService.StopHTTP()
	rpcService.StopIPC()
	rpcService.RpcAPIs = nil
	return nil
}

func (rpcService *RpcService) Receive(context actor.Context) {}

//TODO split big rpc to  small controller （HTTP WS IPC REST
// StartHTTP initializes and starts the HTTP RPC endpoint.
func (rpcService *RpcService) StartRest(endpoint string, restApi rpc.RestDescription) error {
	if !rpcService.Config.RESTEnabled {
		return nil
	}
	go func() {
		mainController := rpc.StartRest(restApi)
		rpcService.RestEndpoint = endpoint
		rpcService.RestController = mainController
	}()
	return nil
}

// StopHTTP terminates the HTTP RPC endpoint.
func (rpcService *RpcService) StopREST() {
	if rpcService.RestController != nil {
		rpcService.RestController.Stop()
		rpcService.RestController = nil
		log.WithField("url", fmt.Sprintf("http://%s", rpcService.HttpEndpoint)).Info("REST endpoint closed")
	}
}

// StartInProc initializes an in-process RPC endpoint.
func (rpcService *RpcService) StartInProc(apis []app.API) error {
	// Register all the APIs exposed by the services
	handler := rpc.NewServer()
	for _, api := range apis {
		if err := handler.RegisterName(api.Namespace, api.Service); err != nil {
			return err
		}
		log.WithField("namespace", api.Namespace).Debug("InProc registered")
	}
	rpcService.inprocHandler = handler
	return nil
}

// StopInProc terminates the in-process RPC endpoint.
func (rpcService *RpcService) StopInProc() {
	if rpcService.inprocHandler != nil {
		rpcService.inprocHandler.Stop()
		rpcService.inprocHandler = nil
	}
}

// StartIPC initializes and starts the IPC RPC endpoint.
func (rpcService *RpcService) StartIPC(apis []app.API) error {
	if !rpcService.Config.IPCEnabled {
		return nil
	}
	if rpcService.IpcEndpoint == "" {
		return nil // IPC disabled.
	}
	listener, handler, err := StartIPCEndpoint(rpcService.IpcEndpoint, apis)
	if err != nil {
		return err
	}
	rpcService.IpcListener = listener
	rpcService.IpcHandler = handler
	log.WithField("url", rpcService.IpcEndpoint).Info("IPC endpoint opened")
	return nil
}

// StopIPC terminates the IPC RPC endpoint.
func (rpcService *RpcService) StopIPC() {
	if rpcService.IpcListener != nil {
		rpcService.IpcListener.Close()
		rpcService.IpcListener = nil

		log.WithField("endpoint", rpcService.IpcEndpoint).Info("IPC endpoint closed")
	}
	if rpcService.IpcHandler != nil {
		rpcService.IpcHandler.Stop()
		rpcService.IpcHandler = nil
	}
}

// StartHTTP initializes and starts the HTTP RPC endpoint.
func (rpcService *RpcService) StartHTTP(endpoint string, apis []app.API, modules []string, cors []string, vhosts []string, timeouts *rpc.HTTPTimeouts) error {
	if !rpcService.Config.HTTPEnabled {
		return nil
	}
	// Short circuit if the HTTP endpoint isn't being exposed
	if endpoint == "" {
		return nil
	}
	listener, handler, err := StartHTTPEndpoint(endpoint, apis, modules, cors, vhosts, rpc.HTTPTimeouts{})
	if err != nil {
		return err
	}
	log.WithField("url", fmt.Sprintf("http://%s", endpoint)).WithField("cors", strings.Join(cors, ",")).WithField("vhosts", strings.Join(vhosts, ",")).Info("HTTP endpoint opened")
	// All listeners booted successfully
	rpcService.HttpEndpoint = endpoint
	rpcService.HttpListener = listener
	rpcService.HttpHandler = handler
	return nil
}

// StopHTTP terminates the HTTP RPC endpoint.
func (rpcService *RpcService) StopHTTP() {
	if rpcService.HttpListener != nil {
		rpcService.HttpListener.Close()
		rpcService.HttpListener = nil

		log.WithField("url", fmt.Sprintf("http://%s", rpcService.HttpEndpoint)).Info("HTTP endpoint closed")
	}
	if rpcService.HttpHandler != nil {
		rpcService.HttpHandler.Stop()
		rpcService.HttpHandler = nil
	}
}

// StartWS initializes and starts the websocket RPC endpoint.
func (rpcService *RpcService) StartWS(endpoint string, apis []app.API, modules []string, wsOrigins []string, exposeAll bool) error {
	if !rpcService.Config.WSEnabled {
		return nil
	}
	// Short circuit if the WS endpoint isn't being exposed
	if endpoint == "" {
		return nil
	}
	listener, handler, err := StartWSEndpoint(endpoint, apis, modules, wsOrigins, exposeAll)
	if err != nil {
		return err
	}
	log.WithField("url", fmt.Sprintf("ws://%s", listener.Addr())).Info("WebSocket endpoint opened")
	// All listeners booted successfully
	rpcService.WsEndpoint = endpoint
	rpcService.WsListener = listener
	rpcService.WsHandler = handler

	return nil
}

// StopWS terminates the websocket RPC endpoint.
func (rpcService *RpcService) StopWS() {
	if rpcService.WsListener != nil {
		rpcService.WsListener.Close()
		rpcService.WsListener = nil

		log.WithField("url", fmt.Sprintf("ws://%s", rpcService.WsEndpoint)).Info("WebSocket endpoint closed")
	}
	if rpcService.WsHandler != nil {
		rpcService.WsHandler.Stop()
		rpcService.WsHandler = nil
	}
}

// setRpc creates an rpc configuration from the set command line flags,
func (rpcService *RpcService) setRpcLog(ctx *cli.Context, homeDir string) {
	rpcService.setIPC(ctx, homeDir)
	rpcService.setHTTP(ctx, homeDir)
	rpcService.setWS(ctx, homeDir)
	rpcService.setRest(ctx, homeDir)
}

// setIPC creates an IPC path configuration from the set command line flags,
// returning an empty string if IPC was explicitly disabled, or the set path.
func (rpcService *RpcService) setIPC(ctx *cli.Context, homeDir string) {
	rpcService.Config.IPCEnabled = true
	if ctx.GlobalBool(IPCDisabledFlag.Name) {
		rpcService.Config.IPCEnabled = false
		return
	}

	checkExclusive(ctx, IPCDisabledFlag, IPCPathFlag)
	if ctx.GlobalIsSet(IPCPathFlag.Name) {
		rpcService.Config.IPCPath = ctx.GlobalString(IPCPathFlag.Name)
	} else {
		rpcService.Config.IPCPath = path.Join(homeDir, DefaultIPCEndpoint(ClientIdentifier))
	}
}

// setHTTP creates the HTTP RPC listener interface string from the set
// command line flags, returning empty if the HTTP endpoint is disabled.
func (rpcService *RpcService) setHTTP(ctx *cli.Context, homeDir string) {
	if !rpcService.Config.HTTPEnabled {
		if ctx.GlobalBool(HTTPEnabledFlag.Name) {
			rpcService.Config.HTTPEnabled = true
		}
	}

	if ctx.GlobalIsSet(HTTPListenAddrFlag.Name) {
		rpcService.Config.HTTPHost = ctx.GlobalString(HTTPListenAddrFlag.Name)
	} else {
		if rpcService.Config.HTTPHost == "" {
			rpcService.Config.HTTPHost = rpc.DefaultHTTPHost
		}
	}

	if ctx.GlobalIsSet(HTTPPortFlag.Name) {
		rpcService.Config.HTTPPort = ctx.GlobalInt(HTTPPortFlag.Name)
	} else {
		if rpcService.Config.HTTPPort == 0 {
			rpcService.Config.HTTPPort = rpc.DefaultHTTPPort
		}
	}

	if ctx.GlobalIsSet(HTTPCORSDomainFlag.Name) {
		rpcService.Config.HTTPCors = splitAndTrim(ctx.GlobalString(HTTPCORSDomainFlag.Name))
	}

	if ctx.GlobalIsSet(HTTPApiFlag.Name) {
		rpcService.Config.HTTPModules = splitAndTrim(ctx.GlobalString(HTTPApiFlag.Name))
	}

	if ctx.GlobalIsSet(HTTPVirtualHostsFlag.Name) {
		rpcService.Config.HTTPVirtualHosts = splitAndTrim(ctx.GlobalString(HTTPVirtualHostsFlag.Name))
	} else {
		if rpcService.Config.HTTPVirtualHosts == nil {
			rpcService.Config.HTTPVirtualHosts = []string{"localhost"}
		}
	}
}

// setHTTP creates the HTTP RPC listener interface string from the set
// command line flags, returning empty if the HTTP endpoint is disabled.
func (rpcService *RpcService) setRest(ctx *cli.Context, homeDir string) {
	if !rpcService.Config.RESTEnabled {
		if ctx.GlobalBool(RESTEnabledFlag.Name) {
			rpcService.Config.RESTEnabled = true
		}
	}

	if ctx.GlobalIsSet(RESTListenAddrFlag.Name) {
		rpcService.Config.RESTHost = ctx.GlobalString(RESTListenAddrFlag.Name)
	} else {
		if rpcService.Config.RESTHost == "" {
			rpcService.Config.RESTHost = rpc.DefaultRestHost
		}
	}

	if ctx.GlobalIsSet(RESTPortFlag.Name) {
		rpcService.Config.RESTPort = ctx.GlobalInt(RESTPortFlag.Name)
	} else {
		if rpcService.Config.RESTPort == 0 {
			rpcService.Config.RESTPort = rpc.DefaultRestPort
		}
	}
}

// setWS creates the WebSocket RPC listener interface string from the set
// command line flags, returning empty if the HTTP endpoint is disabled.
func (rpcService *RpcService) setWS(ctx *cli.Context, homeDir string) {
	if !rpcService.Config.WSEnabled {
		if ctx.GlobalBool(WSEnabledFlag.Name) {
			rpcService.Config.WSEnabled = true
		}
	}

	if ctx.GlobalIsSet(WSListenAddrFlag.Name) {
		rpcService.Config.WSHost = ctx.GlobalString(WSListenAddrFlag.Name)
	} else {
		if rpcService.Config.WSHost == "" {
			rpcService.Config.WSHost = rpc.DefaultWSHost
		}
	}

	if ctx.GlobalIsSet(WSPortFlag.Name) {
		rpcService.Config.WSPort = ctx.GlobalInt(WSPortFlag.Name)
	} else {
		if rpcService.Config.WSPort == 0 {
			rpcService.Config.WSPort = rpc.DefaultWSPort
		}
	}

	if ctx.GlobalIsSet(WSAllowedOriginsFlag.Name) {
		rpcService.Config.WSOrigins = splitAndTrim(ctx.GlobalString(WSAllowedOriginsFlag.Name))
	}

	if ctx.GlobalIsSet(WSApiFlag.Name) {
		rpcService.Config.WSModules = splitAndTrim(ctx.GlobalString(WSApiFlag.Name))
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
		panic(fmt.Sprintf("Flags %v can't be used at the same time", strings.Join(set, ", ")))
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

// DefaultIPCEndpoint returns the IPC path used by default.
func DefaultIPCEndpoint(clientIdentifier string) string {
	if clientIdentifier == "" {
		clientIdentifier = strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
		if clientIdentifier == "" {
			panic("empty executable name")
		}
	}

	return clientIdentifier + ".ipc"
}

func (rpcService *RpcService) DefaultConfig(netType params.NetType) *rpc.RpcConfig {
	return &rpc.RpcConfig{
		HTTPTimeouts: &rpc.DefaultHTTPTimeouts,
	}
}

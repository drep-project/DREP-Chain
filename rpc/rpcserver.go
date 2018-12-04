package rpc

import (
	"fmt"
	"sync"
	"net"
	"os"
	"strings"
	"runtime"
	"path/filepath"

	"BlockChainTest/database"
	"BlockChainTest/accounts"
	"BlockChainTest/node"
	"BlockChainTest/rpc/rest"
	"BlockChainTest/log"
)

const (
	DefaultHTTPHost = "localhost" // Default host interface for the HTTP RPC server
	DefaultHTTPPort = 15645        // Default TCP port for the HTTP RPC server
	DefaultWSHost   = "localhost" // Default host interface for the websocket RPC server
	DefaultWSPort   = 15646        // Default TCP port for the websocket RPC server
	DefaultRestHost = "localhost"  // Default host interface for the REST RPC server
	DefaultRestPort = 156457       // Default TCP port for the REST RPC server
)

type RpcConfig struct {
	// IPCPath is the requested location to place the IPC endpoint. If the path is
	// a simple file name, it is placed inside the data directory (or on the root
	// pipe path on Windows), whereas if it's a resolvable path name (absolute or
	// relative), then that specific path is enforced. An empty path disables IPC.
	IPCPath string `toml:",omitempty"`

	// HTTPHost is the host interface on which to start the HTTP RPC server. If this
	// field is empty, no HTTP API endpoint will be started.
	HTTPHost string `toml:",omitempty"`

	// HTTPPort is the TCP port number on which to start the HTTP RPC server. The
	// default zero value is/ valid and will pick a port number randomly (useful
	// for ephemeral nodes).
	HTTPPort int `toml:",omitempty"`

	// HTTPCors is the Cross-Origin Resource Sharing header to send to requesting
	// clients. Please be aware that CORS is a browser enforced security, it's fully
	// useless for custom HTTP clients.
	HTTPCors []string `toml:",omitempty"`

	// HTTPVirtualHosts is the list of virtual hostnames which are allowed on incoming requests.
	// This is by default {'localhost'}. Using this prevents attacks like
	// DNS rebinding, which bypasses SOP by simply masquerading as being within the same
	// origin. These attacks do not utilize CORS, since they are not cross-domain.
	// By explicitly checking the Host-header, the server will not allow requests
	// made against the server with a malicious host domain.
	// Requests using ip address directly are not affected
	HTTPVirtualHosts []string `toml:",omitempty"`

	// HTTPModules is a list of API modules to expose via the HTTP RPC interface.
	// If the module list is empty, all RPC API endpoints designated public will be
	// exposed.
	HTTPModules []string `toml:",omitempty"`

	// HTTPTimeouts allows for customization of the timeout values used by the HTTP RPC
	// interface.
	HTTPTimeouts HTTPTimeouts

	// WSHost is the host interface on which to start the websocket RPC server. If
	// this field is empty, no websocket API endpoint will be started.
	WSHost string `toml:",omitempty"`

	// WSPort is the TCP port number on which to start the websocket RPC server. The
	// default zero value is/ valid and will pick a port number randomly (useful for
	// ephemeral nodes).
	WSPort int `toml:",omitempty"`

	// WSOrigins is the list of domain to accept websocket requests from. Please be
	// aware that the server can only act upon the HTTP request the client sends and
	// cannot verify the validity of the request header.
	WSOrigins []string `toml:",omitempty"`

	// WSModules is a list of API modules to expose via the websocket RPC interface.
	// If the module list is empty, all RPC API endpoints designated public will be
	// exposed.
	WSModules []string `toml:",omitempty"`

	// WSExposeAll exposes all API modules via the WebSocket RPC interface rather
	// than just the public ones.
	//
	// *WARNING* Only set this if the node is running in a trusted network, exposing
	// private APIs to untrusted users is a major security risk.
	WSExposeAll bool `toml:",omitempty"`


	// HTTPHost is the host interface on which to start the HTTP RPC server. If this
	// field is empty, no HTTP API endpoint will be started.
	RESTHost string `toml:",omitempty"`

	// HTTPPort is the TCP port number on which to start the HTTP RPC server. The
	// default zero value is/ valid and will pick a port number randomly (useful
	// for ephemeral nodes).
	RESTPort int `toml:",omitempty"`

	// HTTPCors is the Cross-Origin Resource Sharing header to send to requesting
	// clients. Please be aware that CORS is a browser enforced security, it's fully
	// useless for custom HTTP clients.
	RESTCors []string `toml:",omitempty"`
}


// IPCEndpoint resolves an IPC endpoint based on a configured value, taking into
// account the set data folders as well as the designated platform we're currently
// running on.
func (c *RpcConfig) IPCEndpoint() string {
	// Short circuit if IPC has not been enabled
	if c.IPCPath == "" {
		return ""
	}
	// On windows we can only use plain top-level pipes
	if runtime.GOOS == "windows" {
		if strings.HasPrefix(c.IPCPath, `\\.\pipe\`) {
			return c.IPCPath
		}
		return `\\.\pipe\` + c.IPCPath
	}
	// Resolve names into the data directory full paths otherwise
	if filepath.Base(c.IPCPath) == c.IPCPath {
		//if c.DataDir == "" {
			return filepath.Join(os.TempDir(), c.IPCPath)
		//}
		//return filepath.Join(c.DataDir, c.IPCPath)
	}
	return c.IPCPath
}

// HTTPEndpoint resolves an HTTP endpoint based on the configured host interface
// and port parameters.
func (c *RpcConfig) HTTPEndpoint() string {
	if c.HTTPHost == "" {
		return ""
	}
	return fmt.Sprintf("%s:%d", c.HTTPHost, c.HTTPPort)
}

// HTTPEndpoint resolves an HTTP endpoint based on the configured host interface
// and port parameters.
func (c *RpcConfig) RestEndpoint() string {
	if c.HTTPHost == "" {
		return ""
	}
	return fmt.Sprintf("%s:%d", c.RESTHost, c.RESTPort)
}

// DefaultHTTPEndpoint returns the HTTP endpoint used by default.
func DefaultHTTPEndpoint() string {
	config := &RpcConfig{HTTPHost: DefaultHTTPHost, HTTPPort: DefaultHTTPPort}
	return config.HTTPEndpoint()
}

// DefaultHTTPEndpoint returns the HTTP endpoint used by default.
func DefaultRestEndpoint() string {
	config := &RpcConfig{HTTPHost: DefaultRestHost, HTTPPort: DefaultRestPort}
	return config.RestEndpoint()
}


// WSEndpoint resolves a websocket endpoint based on the configured host interface
// and port parameters.
func (c *RpcConfig) WSEndpoint() string {
	if c.WSHost == "" {
		return ""
	}
	return fmt.Sprintf("%s:%d", c.WSHost, c.WSPort)
}

// DefaultWSEndpoint returns the websocket endpoint used by default.
func DefaultWSEndpoint() string {
	config := &RpcConfig{WSHost: DefaultWSHost, WSPort: DefaultWSPort}
	return config.WSEndpoint()
}

type RpcServer struct {
	rpcAPIs       []API   // List of APIs currently provided by the node
	inprocHandler *Server // In-process RPC request handler to process the API requests

	ipcEndpoint string       // IPC endpoint to listen at (empty = IPC disabled)
	ipcListener net.Listener // IPC RPC listener socket to serve API requests
	ipcHandler  *Server  // IPC RPC request handler to process the API requests

	httpEndpoint  string       // HTTP endpoint (interface + port) to listen at (empty = HTTP disabled)
	httpWhitelist []string     // HTTP RPC modules to allow through this endpoint
	httpListener  net.Listener // HTTP RPC listener socket to server API requests
	httpHandler   *Server  // HTTP RPC request handler to process the API requests

	wsEndpoint string       // Websocket endpoint (interface + port) to listen at (empty = websocket disabled)
	wsListener net.Listener // Websocket RPC listener socket to server API requests
	wsHandler  *Server  // Websocket RPC request handler to process the API requests
	
	lock sync.RWMutex
	rpcConfig *RpcConfig
}


func NewRpcServer(rpcConfig *RpcConfig)*RpcServer{
	api := API{
		Namespace : "database",
		Version   :"1.0",
		Service  : &database.DataBaseAPI{},
		Public  :  true      ,
	}
	chainApi := API{
		Namespace : "chain",
		Version   :"1.0",
		Service:	&node.ChainApi{},
		Public  :  true      ,
	}
	accountApi := API{
		Namespace : "account",
		Version   :"1.0",
		Service:	&accounts.AccountApi{},
		Public  :  true      ,
	}
	
    return &RpcServer{
		ipcEndpoint: rpcConfig.IPCEndpoint(),
		httpEndpoint:  rpcConfig.HTTPEndpoint(),
		wsEndpoint:  rpcConfig.WSEndpoint(),
		rpcConfig: rpcConfig,
		rpcAPIs:[]API{api,chainApi,accountApi},
    }
}

func (rpcserver *RpcServer) GetIpcHandler()*Server{
     return rpcserver.ipcHandler
}
func (rpcserver *RpcServer) GetHttpHandler()*Server{
	return rpcserver.httpHandler
}
func (rpcserver *RpcServer) GetConfig() *RpcConfig {
	return rpcserver.rpcConfig
}

// startRPC is a helper method to start all the various RPC endpoint during node
// startup. It's not meant to be called at any time afterwards as it makes certain
// assumptions about the state of the node.
func (rpcserver *RpcServer) StartRPC() error {
	// All API endpoints started successfully
	//rpcserver.rpcAPIs = apis
	// Start the various API endpoints, terminating all in case of errors
	if err := rpcserver.startInProc(rpcserver.rpcAPIs); err != nil {
		return err
	}
	if err := rpcserver.startIPC(rpcserver.rpcAPIs); err != nil {
		rpcserver.stopInProc()
		return err
	}
	if err := rpcserver.startHTTP(rpcserver.httpEndpoint, rpcserver.rpcAPIs, rpcserver.rpcConfig.HTTPModules, rpcserver.rpcConfig.HTTPCors, rpcserver.rpcConfig.HTTPVirtualHosts, rpcserver.rpcConfig.HTTPTimeouts); err != nil {
		rpcserver.stopIPC()
		rpcserver.stopInProc()
		return err
	}
	if err := rpcserver.startWS(rpcserver.wsEndpoint, rpcserver.rpcAPIs, rpcserver.rpcConfig.WSModules, rpcserver.rpcConfig.WSOrigins, rpcserver.rpcConfig.WSExposeAll); err != nil {
		rpcserver.stopHTTP()
		rpcserver.stopIPC()
		rpcserver.stopInProc()
		return err
	}
	rest.HttpStart(DefaultRestEndpoint())
	return nil
}

// startInProc initializes an in-process RPC endpoint.
func (rpcserver *RpcServer) startInProc(apis []API) error {
	// Register all the APIs exposed by the services
	handler := NewServer()
	for _, api := range apis {
		if err := handler.RegisterName(api.Namespace, api.Service); err != nil {
			return err
		}
		log.Debug("InProc registered", "namespace", api.Namespace)
	}
	rpcserver.inprocHandler = handler
	return nil
}

// stopInProc terminates the in-process RPC endpoint.
func (rpcserver *RpcServer) stopInProc() {
	if rpcserver.inprocHandler != nil {
		rpcserver.inprocHandler.Stop()
		rpcserver.inprocHandler = nil
	}
}

// startIPC initializes and starts the IPC RPC endpoint.
func (rpcserver *RpcServer) startIPC(apis []API) error {
	if rpcserver.ipcEndpoint == "" {
		return nil // IPC disabled.
	}
	listener, handler, err := StartIPCEndpoint(rpcserver.ipcEndpoint, apis)
	if err != nil {
		return err
	}
	rpcserver.ipcListener = listener
	rpcserver.ipcHandler = handler
	log.Info("IPC endpoint opened", "url", rpcserver.ipcEndpoint)
	return nil
}

// stopIPC terminates the IPC RPC endpoint.
func (rpcserver *RpcServer) stopIPC() {
	if rpcserver.ipcListener != nil {
		rpcserver.ipcListener.Close()
		rpcserver.ipcListener = nil

		log.Info("IPC endpoint closed", "endpoint", rpcserver.ipcEndpoint)
	}
	if rpcserver.ipcHandler != nil {
		rpcserver.ipcHandler.Stop()
		rpcserver.ipcHandler = nil
	}
}

// startHTTP initializes and starts the HTTP RPC endpoint.
func (rpcserver *RpcServer) startHTTP(endpoint string, apis []API, modules []string, cors []string, vhosts []string, timeouts HTTPTimeouts) error {
	// Short circuit if the HTTP endpoint isn't being exposed
	if endpoint == "" {
		return nil
	}
	listener, handler, err := StartHTTPEndpoint(endpoint, apis, modules, cors, vhosts, timeouts)
	if err != nil {
		return err
	}
	log.Info("HTTP endpoint opened", "url", fmt.Sprintf("http://%s", endpoint), "cors", strings.Join(cors, ","), "vhosts", strings.Join(vhosts, ","))
	// All listeners booted successfully
	rpcserver.httpEndpoint = endpoint
	rpcserver.httpListener = listener
	rpcserver.httpHandler = handler

	return nil
}

// stopHTTP terminates the HTTP RPC endpoint.
func (rpcserver *RpcServer) stopHTTP() {
	if rpcserver.httpListener != nil {
		rpcserver.httpListener.Close()
		rpcserver.httpListener = nil

		log.Info("HTTP endpoint closed", "url", fmt.Sprintf("http://%s", rpcserver.httpEndpoint))
	}
	if rpcserver.httpHandler != nil {
		rpcserver.httpHandler.Stop()
		rpcserver.httpHandler = nil
	}
}

// startWS initializes and starts the websocket RPC endpoint.
func (rpcserver *RpcServer) startWS(endpoint string, apis []API, modules []string, wsOrigins []string, exposeAll bool) error {
	// Short circuit if the WS endpoint isn't being exposed
	if endpoint == "" {
		return nil
	}
	listener, handler, err := StartWSEndpoint(endpoint, apis, modules, wsOrigins, exposeAll)
	if err != nil {
		return err
	}
	log.Info("WebSocket endpoint opened", "url", fmt.Sprintf("ws://%s", listener.Addr()))
	// All listeners booted successfully
	rpcserver.wsEndpoint = endpoint
	rpcserver.wsListener = listener
	rpcserver.wsHandler = handler

	return nil
}

// stopWS terminates the websocket RPC endpoint.
func (rpcserver *RpcServer) stopWS() {
	if rpcserver.wsListener != nil {
		rpcserver.wsListener.Close()
		rpcserver.wsListener = nil

		log.Info("WebSocket endpoint closed", "url", fmt.Sprintf("ws://%s", rpcserver.wsEndpoint))
	}
	if rpcserver.wsHandler != nil {
		rpcserver.wsHandler.Stop()
		rpcserver.wsHandler = nil
	}
}

// Stop terminates a running node along with all it's services. In the node was
// not started, an error is returned.
func (rpcserver *RpcServer) Stop() error {
	rpcserver.lock.Lock()
	defer rpcserver.lock.Unlock()
	// Terminate the API, services and the p2p server.
	rpcserver.stopWS()
	rpcserver.stopHTTP()
	rpcserver.stopIPC()
	rpcserver.rpcAPIs = nil
	return nil
}
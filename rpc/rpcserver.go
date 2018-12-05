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
	DefaultRestPort = 15647       // Default TCP port for the REST RPC server
)

type RpcConfig struct {

	// IPCEnabled 
	IPCEnabled bool

	// IPCPath is the requested location to place the IPC endpoint. If the path is
	// a simple file name, it is placed inside the data directory (or on the root
	// pipe path on Windows), whereas if it's a resolvable path name (absolute or
	// relative), then that specific path is enforced. An empty path disables IPC.
	IPCPath string `toml:",omitempty"`

	// HTTPEnabled  
	HTTPEnabled bool
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

	// WSEnabled  
	WSEnabled bool
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

	// RESTEnabled  
	RESTEnabled bool
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
	if c.RESTHost == "" {
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
	RpcAPIs       []API   // List of APIs currently provided by the node
	inprocHandler *Server // In-process RPC request handler to process the API requests

	IpcEndpoint string       // IPC endpoint to listen at (empty = IPC disabled)
	IpcListener net.Listener // IPC RPC listener socket to serve API requests
	IpcHandler  *Server  // IPC RPC request handler to process the API requests

	HttpEndpoint  string       // HTTP endpoint (interface + port) to listen at (empty = HTTP disabled)
	HttpWhitelist []string     // HTTP RPC modules to allow through this endpoint
	HttpListener  net.Listener // HTTP RPC listener socket to server API requests
	HttpHandler   *Server  // HTTP RPC request handler to process the API requests

	WsEndpoint string       // Websocket endpoint (interface + port) to listen at (empty = websocket disabled)
	WsListener net.Listener // Websocket RPC listener socket to server API requests
	WsHandler  *Server  // Websocket RPC request handler to process the API requests
	
	RestEndpoint string       // Websocket endpoint (interface + port) to listen at (empty = websocket disabled)
	RestListener net.Listener // Websocket RPC listener socket to server API requests

	lock sync.RWMutex
	RpcConfig *RpcConfig
}


func NewRpcServer(RpcConfig *RpcConfig)*RpcServer{
	api := API{
		Namespace : "db",
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
		IpcEndpoint: RpcConfig.IPCEndpoint(),
		HttpEndpoint:  RpcConfig.HTTPEndpoint(),
		WsEndpoint:  RpcConfig.WSEndpoint(),
		RestEndpoint: RpcConfig.RestEndpoint(),
		RpcConfig: RpcConfig,
		RpcAPIs:[]API{api,chainApi,accountApi},
    }
}

// startRPC is a helper method to start all the various RPC endpoint during node
// startup. It's not meant to be called at any time afterwards as it makes certain
// assumptions about the state of the node.
func (rpcserver *RpcServer) StartRPC() error {
	// All API endpoints started successfully
	//rpcserver.RpcAPIs = apis
	// Start the various API endpoints, terminating all in case of errors
	if err := rpcserver.StartInProc(rpcserver.RpcAPIs); err != nil {
		return err
	}
	if err := rpcserver.StartIPC(rpcserver.RpcAPIs); err != nil {
		rpcserver.StopInProc()
		return err
	}
	if err := rpcserver.StartHTTP(rpcserver.HttpEndpoint, rpcserver.RpcAPIs, rpcserver.RpcConfig.HTTPModules, rpcserver.RpcConfig.HTTPCors, rpcserver.RpcConfig.HTTPVirtualHosts, rpcserver.RpcConfig.HTTPTimeouts); err != nil {
		rpcserver.StopIPC()
		rpcserver.StopInProc()
		return err
	}
	if err := rpcserver.StartWS(rpcserver.WsEndpoint, rpcserver.RpcAPIs, rpcserver.RpcConfig.WSModules, rpcserver.RpcConfig.WSOrigins, rpcserver.RpcConfig.WSExposeAll); err != nil {
		rpcserver.StopHTTP()
		rpcserver.StopIPC()
		rpcserver.StopInProc()
		return err
	}

	if err := rpcserver.StartREST(rpcserver.RestEndpoint); err != nil {
		rpcserver.StopREST()
		return err
	}
	return nil
}

// StartInProc initializes an in-process RPC endpoint.
func (rpcserver *RpcServer) StartInProc(apis []API) error {
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

// StopInProc terminates the in-process RPC endpoint.
func (rpcserver *RpcServer) StopInProc() {
	if rpcserver.inprocHandler != nil {
		rpcserver.inprocHandler.Stop()
		rpcserver.inprocHandler = nil
	}
}

// StartIPC initializes and starts the IPC RPC endpoint.
func (rpcserver *RpcServer) StartIPC(apis []API) error {
	if !rpcserver.RpcConfig.IPCEnabled {
		return nil
	}
	if rpcserver.IpcEndpoint == "" {
		return nil // IPC disabled.
	}
	listener, handler, err := StartIPCEndpoint(rpcserver.IpcEndpoint, apis)
	if err != nil {
		return err
	}
	rpcserver.IpcListener = listener
	rpcserver.IpcHandler = handler
	log.Info("IPC endpoint opened", "url", rpcserver.IpcEndpoint)
	return nil
}

// StopIPC terminates the IPC RPC endpoint.
func (rpcserver *RpcServer) StopIPC() {
	if rpcserver.IpcListener != nil {
		rpcserver.IpcListener.Close()
		rpcserver.IpcListener = nil

		log.Info("IPC endpoint closed", "endpoint", rpcserver.IpcEndpoint)
	}
	if rpcserver.IpcHandler != nil {
		rpcserver.IpcHandler.Stop()
		rpcserver.IpcHandler = nil
	}
}

// StartHTTP initializes and starts the HTTP RPC endpoint.
func (rpcserver *RpcServer) StartHTTP(endpoint string, apis []API, modules []string, cors []string, vhosts []string, timeouts HTTPTimeouts) error {
	if !rpcserver.RpcConfig.HTTPEnabled {
		return nil
	}
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
	rpcserver.HttpEndpoint = endpoint
	rpcserver.HttpListener = listener
	rpcserver.HttpHandler = handler

	return nil
}

// StopHTTP terminates the HTTP RPC endpoint.
func (rpcserver *RpcServer) StopHTTP() {
	if rpcserver.HttpListener != nil {
		rpcserver.HttpListener.Close()
		rpcserver.HttpListener = nil

		log.Info("HTTP endpoint closed", "url", fmt.Sprintf("http://%s", rpcserver.HttpEndpoint))
	}
	if rpcserver.HttpHandler != nil {
		rpcserver.HttpHandler.Stop()
		rpcserver.HttpHandler = nil
	}
}

// StartWS initializes and starts the websocket RPC endpoint.
func (rpcserver *RpcServer) StartWS(endpoint string, apis []API, modules []string, wsOrigins []string, exposeAll bool) error {
	if !rpcserver.RpcConfig.WSEnabled {
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
	log.Info("WebSocket endpoint opened", "url", fmt.Sprintf("ws://%s", listener.Addr()))
	// All listeners booted successfully
	rpcserver.WsEndpoint = endpoint
	rpcserver.WsListener = listener
	rpcserver.WsHandler = handler

	return nil
}

// StopWS terminates the websocket RPC endpoint.
func (rpcserver *RpcServer) StopWS() {
	if rpcserver.WsListener != nil {
		rpcserver.WsListener.Close()
		rpcserver.WsListener = nil

		log.Info("WebSocket endpoint closed", "url", fmt.Sprintf("ws://%s", rpcserver.WsEndpoint))
	}
	if rpcserver.WsHandler != nil {
		rpcserver.WsHandler.Stop()
		rpcserver.WsHandler = nil
	}
}

// Stop terminates a running node along with all it's services. In the node was
// not started, an error is returned.
func (rpcserver *RpcServer) Stop() error {
	rpcserver.lock.Lock()
	defer rpcserver.lock.Unlock()
	// Terminate the API, services and the p2p server.
	rpcserver.StopWS()
	rpcserver.StopHTTP()
	rpcserver.StopIPC()
	rpcserver.RpcAPIs = nil
	return nil
}


// StartHTTP initializes and starts the HTTP RPC endpoint.
func (rpcserver *RpcServer) StartREST(endpoint string) error {
	if !rpcserver.RpcConfig.RESTEnabled {
		return nil
	}
	listener, err := rest.HttpStart(endpoint)
	if err != nil {
		return err
	}
	rpcserver.RestEndpoint = endpoint
	rpcserver.RestListener = listener
	return nil
}

// StopHTTP terminates the HTTP RPC endpoint.
func (rpcserver *RpcServer) StopREST()  {
	if rpcserver.RestListener != nil {
		rpcserver.RestListener.Close()
		rpcserver.RestListener = nil

		log.Info("REST endpoint closed", "url", fmt.Sprintf("http://%s", rpcserver.HttpEndpoint))
	}
}
package main

import(
	"fmt"
	"strings"
	"BlockChainTest/rpc"
)
// PrivateAdminAPI is the collection of administrative API methods exposed only
// over a secure RPC channel.
type PrivateAdminAPI struct {
	node *Node // Node interfaced by this API
}

// StartRPC starts the HTTP RPC API server.
func (api *PrivateAdminAPI) StartRPC(host *string, port *int, cors *string, apis *string, vhosts *string) (bool, error) {
	api.node.lock.Lock()
	defer api.node.lock.Unlock()

	rpcServer := api.node.rpcServer
	rpcConfig := api.node.nodeConfig.RpcConfig
	if rpcServer.HttpEndpoint != "" {
		return false, fmt.Errorf("HTTP RPC already running on %s", rpcServer.HttpEndpoint)
	}

	if host == nil {
		h := rpc.DefaultHTTPHost
		if rpcConfig.HTTPHost != "" {
			h = rpcConfig.HTTPHost
		}
		host = &h
	}
	if port == nil {
		port = &rpcConfig.HTTPPort
	}

	allowedOrigins := rpcConfig.HTTPCors
	if cors != nil {
		allowedOrigins = nil
		for _, origin := range strings.Split(*cors, ",") {
			allowedOrigins = append(allowedOrigins, strings.TrimSpace(origin))
		}
	}

	allowedVHosts := rpcConfig.HTTPVirtualHosts
	if vhosts != nil {
		allowedVHosts = nil
		for _, vhost := range strings.Split(*host, ",") {
			allowedVHosts = append(allowedVHosts, strings.TrimSpace(vhost))
		}
	}

	modules := rpcServer.HttpWhitelist
	if apis != nil {
		modules = nil
		for _, m := range strings.Split(*apis, ",") {
			modules = append(modules, strings.TrimSpace(m))
		}
	}

	if err := rpcServer.StartHTTP(fmt.Sprintf("%s:%d", *host, *port), rpcServer.RpcAPIs, modules, allowedOrigins, allowedVHosts, rpcConfig.HTTPTimeouts); err != nil {
		return false, err
	}
	return true, nil
}

// StopRPC terminates an already running HTTP RPC API endpoint.
func (api *PrivateAdminAPI) StopRPC() (bool, error) {
	api.node.lock.Lock()
	defer api.node.lock.Unlock()
	rpcServer := api.node.rpcServer
	if api.node.rpcServer.HttpEndpoint == "" {
		return false, fmt.Errorf("HTTP RPC not running")
	}
	rpcServer.StopHTTP()
	return true, nil
}

// StartWS starts the websocket RPC API server.
func (api *PrivateAdminAPI) StartWS(host *string, port *int, allowedOrigins *string, apis *string) (bool, error) {
	api.node.lock.Lock()
	defer api.node.lock.Unlock()

	if api.node.rpcServer.WsHandler != nil {
		return false, fmt.Errorf("WebSocket RPC already running on %s", api.node.rpcServer.WsHandler )
	}
	rpcServer := api.node.rpcServer
	rpcConfig := api.node.nodeConfig.RpcConfig
	if host == nil {
		h := rpc.DefaultWSHost
		if rpcConfig.WSHost != "" {
			h = rpcConfig.WSHost
		}
		host = &h
	}
	if port == nil {
		port = &rpcConfig.WSPort
	}

	origins := rpcConfig.WSOrigins
	if allowedOrigins != nil {
		origins = nil
		for _, origin := range strings.Split(*allowedOrigins, ",") {
			origins = append(origins, strings.TrimSpace(origin))
		}
	}

	modules :=rpcConfig.WSModules
	if apis != nil {
		modules = nil
		for _, m := range strings.Split(*apis, ",") {
			modules = append(modules, strings.TrimSpace(m))
		}
	}

	if err := rpcServer.StartWS(fmt.Sprintf("%s:%d", *host, *port), rpcServer.RpcAPIs, modules, origins, rpcConfig.WSExposeAll); err != nil {
		return false, err
	}
	return true, nil
}

// StopWS terminates an already running websocket RPC API endpoint.
func (api *PrivateAdminAPI) StopWS() (bool, error) {
	api.node.lock.Lock()
	defer api.node.lock.Unlock()
	rpcServer := api.node.rpcServer
	if rpcServer.WsHandler == nil {
		return false, fmt.Errorf("WebSocket RPC not running")
	}
	rpcServer.StopWS()
	return true, nil
}

package main

import (
	"BlockChainTest/network"
	"BlockChainTest/node"
	"BlockChainTest/processor"
	"BlockChainTest/rpc"
	"BlockChainTest/store"
	"sync"
	"os"
	
	"strings"
	"path/filepath"
)

type Node struct {
	lock sync.RWMutex
	rpcServer *rpc.RpcServer
	StartComplete  chan struct{}
	stopChanel   chan struct{}
}


func (n *Node) Start() {
	n.StartComplete =  make(chan struct{},1)
	go func (){
		cancel := make(chan struct{})
		network.Start(func(peer *network.Peer, t int, msg interface{}) {
			p := processor.GetInstance()
			if msg != nil {
				p.Process(peer, t, msg)
			}
		}, store.GetPort())

		defautConfig := &rpc.RpcConfig{
			IPCPath:DefaultIPCEndpoint(clientIdentifier),
			HTTPHost : rpc.DefaultHTTPHost ,
			HTTPPort : rpc.DefaultHTTPPort,
			WSHost : rpc.DefaultWSHost,
			WSPort : rpc.DefaultWSPort,
			HTTPVirtualHosts : []string{"localhost"},
		}

		n.rpcServer = rpc.NewRpcServer(defautConfig)
		n.rpcServer.StartRPC()
		processor.GetInstance().Start()
		node.GetNode().Start()
		n.StartComplete  <- struct{}{}
		for {
			select {
			case <-cancel   :
				 //todo ctl+c
				n.stopChanel <- struct{}{}
				return
			}
		}
	}()
}
func (n *Node) Wait() {
	for {
		select {
		case <-n.stopChanel :
			return
		}
	}
}

func (n *Node) Attach() (*rpc.Client, error) {
	n.lock.RLock()
	defer n.lock.RUnlock()

	return rpc.DialInProc(n.GetIpcHandler()), nil
}



func (n *Node) GetIpcHandler()*rpc.Server{
	return n.rpcServer.GetIpcHandler()
}
func (n *Node) GetHttpHandler()*rpc.Server{
	return n.rpcServer.GetHttpHandler()
}

func DefaultDataDir() string {
	return "dataDir"
}

func DefaultLogDir() string {
	return "logDir"
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
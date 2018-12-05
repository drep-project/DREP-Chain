package main

import (

	"sync"
	"os"
	"strings"
	"path/filepath"

	"BlockChainTest/network"
	"BlockChainTest/node"
	"BlockChainTest/processor"
	"BlockChainTest/cli/drepcli/utils"
	"BlockChainTest/rpc"
)

type Node struct {
	lock sync.RWMutex
	rpcServer *rpc.RpcServer
	nodeConfig *nodeConfig
	StartComplete  chan struct{}
	stopChanel   chan struct{}
}

func NewNode(nCfg *nodeConfig) Node{
	return Node{
		nodeConfig:nCfg,
	}
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
		}, network.DefaultPort())
		n.rpcServer = rpc.NewRpcServer(&n.nodeConfig.RpcConfig)
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

	return rpc.DialInProc(n.rpcServer.IpcHandler), nil
}

func DefaultDataDir() string {
	return utils.AppDataDir("drep", false)
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
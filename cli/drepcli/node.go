package main

import (
	"sync"

	"BlockChainTest/rpc"
	"BlockChainTest/node"
	"BlockChainTest/bean"
	"BlockChainTest/config"
	"BlockChainTest/network"
	"BlockChainTest/store"
	"BlockChainTest/accounts"
	"BlockChainTest/database"
	"BlockChainTest/processor"
)

type Node struct {
	lock sync.RWMutex
	rpcServer *rpc.RpcServer
	nodeConfig *config.NodeConfig
	StartComplete  chan struct{}
	stopChanel   chan struct{}
}

func NewNode(nCfg *config.NodeConfig) Node{
	return Node{
		nodeConfig:nCfg,
	}
}

func (n *Node) Start() {
	n.StartComplete =  make(chan struct{},1)
	go func (){
		cancel := make(chan struct{})

		database.InitDataBase(n.nodeConfig)
		store.InitState(n.nodeConfig)

		network.Start(func(peer *bean.Peer, t int, msg interface{}) {
			p := processor.GetInstance()
			if msg != nil {
				p.Process(peer, t, msg)
			}
		}, store.GetPort())

		n.rpcServer = rpc.NewRpcServer(n.GetApis(),&n.nodeConfig.RpcConfig)
		n.rpcServer.StartRPC()
		processor.GetInstance().Start()
		node.GetNode().Start(n.nodeConfig)

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
func (n *Node) GetApis() []rpc.API{
	api := rpc.API{
		Namespace : "db",
		Version   :"1.0",
		Service  : &database.DataBaseAPI{},
		Public  :  true      ,
	}
	chainApi := rpc.API{
		Namespace : "chain",
		Version   :"1.0",
		Service:	&node.ChainApi{},
		Public  :  true      ,
	}
	accountApi := rpc.API{
		Namespace : "account",
		Version   :"1.0",
		Service:	&accounts.AccountApi{
			KeyStoreDir : n.nodeConfig.Keystore,
			ChainId : n.nodeConfig.ChainId,
		},
		Public  :  true      ,
	}
	return []rpc.API{api, chainApi, accountApi}
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
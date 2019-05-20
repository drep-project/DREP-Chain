package service

import (
	"github.com/drep-project/dlog"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/network/p2p"
	"github.com/drep-project/drep-chain/network/p2p/enode"
	p2pTypes "github.com/drep-project/drep-chain/network/types"
	"gopkg.in/urfave/cli.v1"
	"path"
)

const (
	MaxConnections = 4000
)

type P2pService struct {
	prvKey   *secp256k1.PrivateKey
	apis     []app.API
	Config   *p2pTypes.P2pConfig
	outQuene chan *outMessage //消息发出去前，要进入此缓存中
	quit     chan struct{}
	server   *p2p.Server //底层p2p管理器
}

type outMessage struct {
	w       p2p.MsgWriter
	msgType uint64
	Msg     interface{}
	done    chan error
}

func (p2pService *P2pService) Name() string {
	return "p2p"
}

func (p2pService *P2pService) Api() []app.API {
	return p2pService.apis
}

func (p2pService *P2pService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{}
}

func (p2pService *P2pService) Init(executeContext *app.ExecuteContext) error {
	// config
	p2pService.Config = p2pTypes.DefaultP2pConfig
	err := executeContext.UnmashalConfig(p2pService.Name(), p2pService.Config)
	if err != nil {
		dlog.Error("p2pService init err", "err", err)
		return err
	}

	p2pService.Config.DataDir = executeContext.CommonConfig.HomeDir
	p2pService.outQuene = make(chan *outMessage, MaxConnections*2)

	if p2pService.Config.PrivateKey == nil {
		p2pService.Config.PrivateKey = p2pService.Config.GeneratePrivateKey()
	}
	//n.serverConfig.Name = n.config.NodeName()
	//n.serverConfig.Logger = n.log
	//if n.serverConfig.StaticNodes == nil {
	//	n.serverConfig.StaticNodes = n.config.StaticNodes()
	//}
	//if n.serverConfig.TrustedNodes == nil {
	//	n.serverConfig.TrustedNodes = n.config.TrustedNodes()
	//}
	//if p2pService.Config.NodeDatabase == "" {
	p2pService.Config.NodeDatabase = path.Join(executeContext.CommonConfig.HomeDir, "drepnode","peersnode")
	//}

	p2pService.server = &p2p.Server{
		Config: p2pService.Config.Config,
	}

	p2pService.apis = []app.API{
		app.API{
			Namespace: "p2p",
			Version:   "1.0",
			Service: &P2PApi{
				p2pService: p2pService,
			},
			Public: true,
		},
	}
	return nil
}

func (p2pService *P2pService) AddProtocols(protocols []p2p.Protocol) {
	p2pService.server.ProtocolsBlockChan = append(p2pService.server.ProtocolsBlockChan, protocols[:len(protocols)]...)
}

func (p2pService *P2pService) Start(executeContext *app.ExecuteContext) error {
	p2pService.server.Start()
	go p2pService.sendMessageRoutine()
	return nil
}

func (p2pService *P2pService) Stop(executeContext *app.ExecuteContext) error {
	if p2pService.server == nil {
		return nil
	}
	p2pService.server.Stop()
	if p2pService.quit != nil {
		close(p2pService.quit)
	}

	return nil
}

func (p2pService *P2pService) SendAsync(w p2p.MsgWriter, msgType uint64, msg interface{}) chan error {
	done := make(chan error, 1)
	p2pService.outQuene <- &outMessage{w: w, Msg: msg, msgType: msgType, done: done}
	return done
}

func (p2pService *P2pService) Send(rw p2p.MsgWriter, msgType uint64, msg interface{}) error {
	done := make(chan error, 1)
	p2pService.outQuene <- &outMessage{w: rw, Msg: msg, msgType: msgType, done: done}
	return <-done
}

func (p2pService *P2pService) sendMessageRoutine() {
	for {
		select {
		case outMsg := <-p2pService.outQuene:
			//消息插入到输出队列的后，网络可能出现立即不通的情况。此时消息应该被丢弃。
			go func() {
				err := p2pService.sendMessage(outMsg) //outMsg.execute()
				if err != nil {
					dlog.Error("p2p send msg err", "msg", outMsg.msgType, "err", err.Error())
				}
				select {
				case outMsg.done <- err:
				default:
				}
			}()

		case <-p2pService.quit:
			return
		}
	}
}

func (p2pService *P2pService) sendMessage(outMessage *outMessage) error {
	return p2p.Send(outMessage.w, (uint64)(outMessage.msgType), outMessage.Msg)
}

func (p2pService *P2pService) Peers() ([]*p2p.Peer) {
	peers := p2pService.server.Peers()

	return peers
}

func (p2pService *P2pService) AddPeer(nodeUrl string) error {
	n := enode.Node{}
	err := n.UnmarshalText([]byte(nodeUrl))

	if err == nil {
		p2pService.server.AddPeer(&n)
	} else {
		dlog.Error("add peer", "err", err)
	}
	return err
}

// nodeUrl："enode://e1b2f83b7b0f5845cc74ca12bb40152e520842bbd0597b7770cb459bd40f109178811ebddd6d640100cdb9b661a3a43a9811d9fdc63770032a3f2524257fb62d@192.168.74.1:55555"
func (p2pService *P2pService) RemovePeer(nodeUrl string) {
	n := enode.Node{}
	err := n.UnmarshalText([]byte(nodeUrl))

	if err == nil {
		p2pService.server.RemovePeer(&n)
	} else {
		dlog.Error("remove peer", "err", err)
	}
}

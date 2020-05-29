package service

import (
	"path"

	"github.com/drep-project/DREP-Chain/app"
	"github.com/drep-project/DREP-Chain/crypto/secp256k1"
	"github.com/drep-project/DREP-Chain/network/p2p"
	"github.com/drep-project/DREP-Chain/network/p2p/enode"
	p2pTypes "github.com/drep-project/DREP-Chain/network/types"
	"gopkg.in/urfave/cli.v1"
)

const (
	MaxConnections = 4000
)

type P2pService struct {
	prvKey   *secp256k1.PrivateKey
	apis     []app.API
	Config   *p2pTypes.P2pConfig
	outQuene chan *outMessage //Before the message is sent, it enters this cache
	quit     chan struct{}
	server   *p2p.Server //The underlying p2p manager
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

func NewP2pService(config *p2pTypes.P2pConfig, homeDir string) *P2pService {
	p2pService := &P2pService{}
	// config
	p2pService.Config = p2pTypes.DefaultP2pConfig

	p2pService.Config.DataDir = homeDir
	p2pService.outQuene = make(chan *outMessage, MaxConnections*2)

	if p2pService.Config.PrivateKey == nil {
		p2pService.Config.PrivateKey = p2pService.Config.GeneratePrivateKey()
	}

	p2pService.Config.NodeDatabase = path.Join(homeDir, "drepnode", "peersnode")

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
	return p2pService
}

func (p2pService *P2pService) Init(executeContext *app.ExecuteContext) error {
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
	p2pService.Config.NodeDatabase = path.Join(executeContext.CommonConfig.HomeDir, "drepnode", "peersnode")
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
			//Immediately after the message is inserted into the output queue, the network may become disconnected. At this point the message should be discarded.
			go func() {
				err := p2pService.sendMessage(outMsg) //outMsg.execute()
				if err != nil {
					log.WithField("msg", outMsg.msgType).WithField("err", err).Error("p2p send msg err")
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

func (p2pService *P2pService) Peers() []*p2p.Peer {
	return p2pService.server.Peers()
}

func (p2pService *P2pService) AddPeer(nodeUrl string) error {
	n := enode.Node{}
	err := n.UnmarshalText([]byte(nodeUrl))

	if err == nil {
		p2pService.server.AddPeer(&n)
	} else {
		log.WithField("err", err).Error("add peer")
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
		log.WithField("err", err).Error("remove peer")
	}
}

//func (p2pService *P2pService) SubscribeEvents(ch chan *p2p.PeerEvent) event.Subscription {
//	return p2pService.server.SubscribeEvents(ch)
//}

func (p2pService *P2pService) LocalNode() *enode.Node {
	return p2pService.server.LocalNode()
}

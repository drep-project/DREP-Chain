package service

import (
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/dlog"
	p2pComponent "github.com/drep-project/drep-chain/network/component"
	"github.com/drep-project/drep-chain/network/component/nat"
	p2pTypes "github.com/drep-project/drep-chain/network/types"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"encoding/json"
)

const (
	MaxLivePeer = 200
	MaxDeadPeer = 200
	MaxConnections = 4000
	UPnPStart  = false
)

var (
	DefaultP2pConfig = &p2pTypes.P2pConfig{
		ListerAddr :"0.0.0.0",
		Port: 55555,
		BootNodes:[]p2pTypes.BootNode{},
	}
)

type P2pService struct {
	prvKey *secp256k1.PrivateKey
	livePeer []*p2pTypes.Peer
	deadPeer []*p2pTypes.Peer
	apis []app.API
	Router *p2pTypes.MessageRouter
	Config *p2pTypes.P2pConfig

	outQuene chan *outMessage
	inQuene chan *p2pTypes.RouteIn

	tryTimer *time.Ticker
	connectCount int32
	pid *actor.PID
	peerOpLock sync.RWMutex
	quit chan struct{}
}

type outMessage struct {
	Peer *p2pTypes.Peer
	Msg  interface{}
	done chan error
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

func (p2pService *P2pService) P2pMessages() map[int]interface{} {
	return map[int]interface{}{
		p2pTypes.MsgTypePing : p2pTypes.Ping{},
		p2pTypes.MsgTypePong : p2pTypes.Pong{},
	}
}

func (p2pService *P2pService) Init(executeContext *app.ExecuteContext) error {
	p2pMessages, err := executeContext.GetMessages()
	if err != nil {
		return err
	}
	for msgType, msgInstance := range p2pMessages {
		err := p2pComponent.RegisterMap(msgType, msgInstance)
		if err != nil {
			return err
		}
	}
	// config
	p2pService.Config = DefaultP2pConfig
	err = executeContext.UnmashalConfig(p2pService.Name(), p2pService.Config)
	if err != nil {
		return err
	}

	p2pService.prvKey, err = secp256k1.GeneratePrivateKey(nil)
	if err != nil {
		//TODO shoud never occur
		dlog.Error("generate private key error ", "Reason", err)
		return err
	}
	p2pService.livePeer = []*p2pTypes.Peer{}
	p2pService.deadPeer = []*p2pTypes.Peer{}
	p2pService.inQuene = make(chan *p2pTypes.RouteIn,MaxConnections*2)
	p2pService.outQuene = make(chan *outMessage,MaxConnections*2)
	props := actor.FromProducer(func() actor.Actor {
		return p2pService
	})

	pid, err := actor.SpawnNamed(props, "peer_message")
	if err != nil {
		panic(err)
	}
	p2pService.Router = p2pTypes.NewMsgRouter(p2pService.inQuene)
	p2pService.Router.RegisterMsgHandler(p2pTypes.MsgTypePing,pid)
	p2pService.Router.RegisterMsgHandler(p2pTypes.MsgTypePong,pid)
	p2pService.pid = pid

	p2pService.apis =  []app.API{
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

func (p2pService *P2pService) Start(executeContext *app.ExecuteContext) error {
	p2pService.initBootNodes()
	go p2pService.receiveRoutine()
	go p2pService.sendMessageRoutine()
	go p2pService.recoverDeadPeer()
	return nil
}

func (p2pService *P2pService) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

func (p2pService *P2pService) initBootNodes(){
	//init safe
	for _, bootNode := range p2pService.Config.BootNodes {
		if p2pService.isLocalIp(bootNode.IP) {
			continue
		}

		p2pService.AddPeer(bootNode.IP)
	}
}

func (p2pService *P2pService) receiveRoutine(){
	//room for modification addr := &net.TCPAddr{IP: net.ParseIP("x.x.x.x"), Port: receiver.listeningPort()}
	addr := &net.TCPAddr{Port: p2pService.Config.Port}
	if UPnPStart {
		nat.Map("tcp", p2pService.Config.Port, p2pService.Config.Port, "drep nat")
	}

	listener, err := net.ListenTCP("tcp", addr)
	dlog.Debug("P2p Service started", "addr", listener.Addr())
	if err != nil {
		dlog.Info("error", err)
		return
	}

	for {
		dlog.Info("start listen", "port", p2pService.Config.Port)
		conn, err := listener.AcceptTCP()
		dlog.Info("listen from ", "accept address", conn.RemoteAddr())
		if err != nil {
			continue
		}

		connChanels := make(chan *net.TCPConn,MaxConnections)
		go func(){
			for{
				conn, err := listener.AcceptTCP()
				if err != nil {
					continue
				}
				connChanels <- conn
			}
		}()

		for{
			select {
			case conn := <- connChanels:
				go func(connTemp *net.TCPConn){
					addr := connTemp.RemoteAddr().String()
					msg, msgType, pubkey, err := p2pService.waitForMessage(connTemp)
					if err != nil {
						if err.Error() == "no msg"{
							return
						}else{
							dlog.Debug("receive message error ","ErrMessage", err.Error())
							return
						}
					}
					peer, err := p2pService.preProcessReq(addr, pubkey)
					if err != nil {
						return
					}
					p2pService.inQuene <- &p2pTypes.RouteIn{
						Type: msgType,
						Peer: peer,
						Detail: msg,
					}
				}(conn)
			}
		}
	}
}

func (p2pService *P2pService) preProcessReq(addr string, pk *secp256k1.PublicKey) (*p2pTypes.Peer, error){
	ipPort := strings.Split(addr,":")
	if p2pService.isLocalIp(ipPort[0]) {
		return nil, errors.New("not allow local ip")
	}
	livePeer := p2pService.GetPeer(ipPort[0])
	if livePeer == nil {
		deadPeer := p2pService.selectDeadPeer(ipPort[0])
		if deadPeer != nil {
			deadPeer.Conn.ReStart()
			p2pService.addPeer(deadPeer)
			livePeer = deadPeer
		}else {
			livePeer = p2pTypes.NewPeer( ipPort[0], p2pTypes.DefaultPort, p2pService.handError, p2pService.sendPing) // //no way to find port
			p2pService.addPeer(livePeer)
		}
	}
	return livePeer, nil
}

func (p2pService *P2pService) waitForMessage(conn *net.TCPConn)(interface{}, int, *secp256k1.PublicKey, error){
	defer conn.Close()
	sizeBytes, err := p2pService.receiveMessageInternal(conn, 4)
	size := (int(sizeBytes[0]) << 24) + (int(sizeBytes[1]) << 16) + (int(sizeBytes[2]) << 8) + int(sizeBytes[3])
	if size == 0 {
		return nil, 0, nil, errors.New("no msg")
	}
	bytes, err := p2pService.receiveMessageInternal(conn, size)
	if err != nil {
		return nil, 0, nil, errors.New("fail to read message ")
	}
	//addr := conn.RemoteAddr().String()
	//log.Debug("receive msg", "Addr", conn.RemoteAddr().String(),"Content", string(bytes))
	return p2pComponent.Deserialize(bytes)
}

func (p2pService *P2pService) receiveMessageInternal(conn net.Conn, size int) ([]byte, error) {
	bytes := make([]byte, size)
	offset := 0
	for offset < size {
		n, err := conn.Read(bytes[offset:])
		offset += n
		if err == io.EOF {
			return bytes, nil
		} else if err != nil {
			return nil, err
		}
	}
	return bytes, nil
}

func (p2pService *P2pService) handPing(peer *p2pTypes.Peer, ping *p2pTypes.Ping){
	p2pService.SendAsync(peer,&p2pTypes.Pong{})
}

func (p2pService *P2pService) handPong(peer *p2pTypes.Peer, pong *p2pTypes.Pong){
	select {
	case peer.Conn.PongTimeoutCh <- false:
	default:
	}
}

func (p2pService *P2pService) handError(peer *p2pTypes.Peer, err error){
	if err != nil {
		if pErr,ok := err.(*p2pTypes.PeerError);ok {
			dlog.Error(pErr.Error())
			peer.Conn.Stop()
			p2pService.addDeadPeer(peer)
		}
	}
}


func (p2pService *P2pService) sendPing(peer *p2pTypes.Peer){
	p2pService.SendAsync(peer, &p2pTypes.Ping{})
}

func (p2pService *P2pService) SendAsync(peer *p2pTypes.Peer, msg interface{}) chan error{
	done := make(chan error,1)
	p2pService.outQuene <-  &outMessage{Peer: peer, Msg:msg,done:done}
	return done
}

func (p2pService *P2pService) Send(peer *p2pTypes.Peer, msg interface{}) error{
	done := make(chan error,1)
	p2pService.outQuene <-  &outMessage{Peer: peer, Msg:msg,done:done}
	return <-done
}

func (p2pService *P2pService) Broadcast(msg interface{}){
	for _, peer := range p2pService.livePeer {
		p2pService.outQuene <-  &outMessage{Peer: peer, Msg:msg}
	}
}

func (p2pService *P2pService) sendMessageRoutine(){
	for {
		select {
		case  outMsg := <-p2pService.outQuene:
			go func() {
				err := p2pService.sendMessage(outMsg) //outMsg.execute()
				if err != nil{
					//dead peer
					p2pService.handError(outMsg.Peer,p2pTypes.NewPeerError(err))
					dlog.Error("", "MSG",err.Error())
				}
				select {
				case outMsg.done <- err:
				default:
				}
			}()
		case <- p2pService.quit:
			return
		}
	}
}

func (p2pService *P2pService) sendMessage(outMessage *outMessage) error {
	message, err := p2pComponent.Serialize(outMessage.Msg, p2pService.prvKey)
	if err != nil {
		dlog.Info("error during cipher:", "reason", err)
		return &common.DataError{MyError:common.MyError{Err:err}}
	}
	d, err := time.ParseDuration("3s")
	if err != nil {
		dlog.Error(err.Error())
		return &common.DefaultError{}
	}
	var conn net.Conn
	for i := 0; i <= 2; i++ {
		conn, err = net.DialTimeout("tcp", outMessage.Peer.GetAddr(), d)
		if err == nil {
			break
		} else {
			dlog.Info(fmt.Sprintf("%T %v\n", err, err))
			if ope, ok := err.(*net.OpError); ok {
				dlog.Info(strconv.FormatBool(ope.Timeout()), ope)
			}
			dlog.Info("Retry after 2s")
			time.Sleep(2 * time.Second)
		}
	}
	if err != nil {
		dlog.Info(fmt.Sprintf("%T %v\n", err, err))
		if ope, ok := err.(*net.OpError); ok {
			dlog.Info(strconv.FormatBool(ope.Timeout()), ope)
			if ope.Timeout() {
				return &common.TimeoutError{MyError:common.MyError{Err:ope}}
			} else {
				return &common.ConnectionError{MyError:common.MyError{Err:ope}}
			}
		}
	}
	defer conn.Close()
	now := time.Now()
	d2, err := time.ParseDuration("5s")
	if err != nil {
		dlog.Error(err.Error())
		return &common.DefaultError{}
	} else {
		conn.SetDeadline(now.Add(d2))
	}
	if bytes, err := json.Marshal(message); err == nil {
		size := len(bytes)
		sizeBytes := make([]byte, 4)
		sizeBytes[0] = byte((size & 0xFF000000) >> 24)
		sizeBytes[1] = byte((size & 0x00FF0000) >> 16)
		sizeBytes[2] = byte((size & 0x0000FF00) >> 8)
		sizeBytes[3] = byte(size & 0x000000FF)
		if err := p2pService.sendMessageInternal(conn, sizeBytes); err != nil {
			return &common.TransmissionError{MyError: common.MyError{Err: err}}
		}
		//dlog.Debug("send message", "IP", conn.RemoteAddr(), "Content", string(bytes))
		if err := p2pService.sendMessageInternal(conn, bytes); err != nil {
			dlog.Error("Send error ", "Msg", err)
			return &common.TransmissionError{MyError: common.MyError{Err: err}}
		} else {
			return nil
		}
	} else {
		return &common.DataError{MyError:common.MyError{Err:err}}
	}
}

func (p2pService *P2pService) sendMessageInternal(conn net.Conn, bytes []byte) error {
	offset := 0
	size := len(bytes)
	for offset < size {
		if num, err := conn.Write(bytes[offset:]); err == nil {
			offset += num
		} else {
			return err
		}
	}
	return nil
}

// TODO p2p operate must be consider more details   1) less lock time； 2）fast query  3）correct peer and deadpeer state
func (p2pService *P2pService) recoverDeadPeer(){
	p2pService.tryTimer = time.NewTicker(time.Second * 30)
	for {
		select  {
		case  <-p2pService.tryTimer.C:
			p2pService.peerOpLock.Lock()
			tryPeerCount := 0
			if len(p2pService.deadPeer) < 40 { //TODO   MAXPEER * RATE  200*0.2
				tryPeerCount = len(p2pService.deadPeer)
			} else {
				tryPeerCount = len(p2pService.deadPeer)/5 //RATE
			}
			tryPeer :=  []*p2pTypes.Peer{}
			for i :=0; i < tryPeerCount; i++ {
				tryPeer = append(tryPeer, p2pService.deadPeer[i])
			}
			p2pService.peerOpLock.Unlock()
			for _, deadPeer := range tryPeer {
				if deadPeer.Conn.Connect() {
					deadPeer.Conn.ReStart()
					p2pService.addPeer(deadPeer)
					dlog.Trace("try to connect peer success", "Addr", deadPeer.GetAddr())
				}else{
					deadPeer.Conn.Stop()
					p2pService.addDeadPeer(deadPeer)
					dlog.Trace("try to connect peer fail", "Addr", deadPeer.GetAddr())
				}
			}
		}
	}
}

func (p2pService *P2pService) GetRouter() (*p2pTypes.MessageRouter) {
	return  p2pService.Router
}

func (p2pService *P2pService) Peers()([]*p2pTypes.Peer){
	return p2pService.livePeer
}

func (p2pService *P2pService) GetPeer(ip string)(*p2pTypes.Peer){
	for _,peer := range p2pService.livePeer {
		if peer.Ip == ip {
			return peer
		}
	}
	return nil
}

func (p2pService *P2pService) selectDeadPeer(ip string)(*p2pTypes.Peer){
	for _,peer := range p2pService.deadPeer {
		if peer.Ip == ip {
			return peer
		}
	}
	return nil
}

func (p2pService *P2pService) AddPeer(addr string) error {
	port := p2pTypes.DefaultPort
	ip := addr
	if strings.Contains(addr, ":") {
		ipPort := strings.Split(addr, ":")

		var err error
		ip = ipPort[0]
		port, err = strconv.Atoi(ipPort[1])
		if err != nil {
			return err
		}
	}
	if p2pService.isLocalIp(ip) {
		return nil
	}
	peer := p2pTypes.NewPeer(ip, port, p2pService.handError, p2pService.sendPing)
	if peer.Conn.Connect() {
		peer.Conn.Start()
		p2pService.addPeer(peer)
		return nil
	} else{
		p2pService.addDeadPeer(peer)
		return errors.New("cant not conntect ip" + ip + ":" + strconv.FormatInt(int64(port), 10))
	}
}

func (p2pService *P2pService) RemovePeer(addr string)  {
	ip := addr
	if strings.Contains(addr, ":") {
		ipPort := strings.Split(addr, ":")
		ip = ipPort[0]
	}

	p2pService.peerOpLock.Lock()
	defer p2pService.peerOpLock.Unlock()

	for index, peer := range p2pService.livePeer {
		if peer.Ip ==  ip {
			p2pService.livePeer = append(p2pService.livePeer[0:index], p2pService.livePeer[index+1:len(p2pService.livePeer)]...)
			break
		}
	}

	for index, peer := range p2pService.deadPeer {
		if peer.Ip ==  ip {
			p2pService.deadPeer = append(p2pService.deadPeer[0:index], p2pService.deadPeer[index+1:len(p2pService.deadPeer)]...)
			break
		}
	}
}

func (p2pService *P2pService) addPeer(peer *p2pTypes.Peer){
	p2pService.peerOpLock.Lock()
	defer p2pService.peerOpLock.Unlock()

	if len(p2pService.livePeer) > MaxLivePeer {
		return
	}
	p2pService.livePeer = append(p2pService.livePeer, peer)

	index := p2pService.indexPeer(p2pService.deadPeer,peer)
	if index > -1 {
		p2pService.deadPeer = append(p2pService.deadPeer[0:index], p2pService.deadPeer[index+1:len(p2pService.deadPeer)]...)
	}
}

func (p2pService *P2pService) addDeadPeer(peer *p2pTypes.Peer){
	p2pService.peerOpLock.Lock()
	defer p2pService.peerOpLock.Unlock()

	index :=  p2pService.indexPeer(p2pService.livePeer,peer)
	if index > -1 {
		p2pService.livePeer = append(p2pService.livePeer[0:index], p2pService.livePeer[index+1:len(p2pService.livePeer)]...)
	}


	index = p2pService.indexPeer(p2pService.deadPeer,peer)
	if index == -1 {
		if len(p2pService.deadPeer) > MaxDeadPeer {
			//remove top peer
			p2pService.deadPeer = p2pService.deadPeer[1:len(p2pService.deadPeer)]
		}
		p2pService.deadPeer = append(p2pService.deadPeer,peer)
	}
}

func (p2pService *P2pService) indexPeer(peers []*p2pTypes.Peer,peer *p2pTypes.Peer) int {
	for index, _peer := range  peers {
		if _peer == peer {
			return index
		}
	}
	return -1
}

func (p2pService *P2pService) isLocalIp(ip string) bool{
	addrs, err := net.InterfaceAddrs()
	if err != nil{
		return false
	}
	for _, value := range addrs{
		if ipnet, ok := value.(*net.IPNet); ok{
			if ipnet.IP.String() == ip {
				return  true
			}
		}
	}
	return false
}



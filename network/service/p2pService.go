package service

import (
	"encoding/json"
	"fmt"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/drep-project/drep-chain/app"
	"github.com/drep-project/drep-chain/common"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/drep-chain/log"
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
)

const (
	MaxLivePeer = 200
	MaxDeadPeer = 200
	MaxConnections = 4000
	UPnPStart  = false
)

type P2pService struct {
	prvKey *secp256k1.PrivateKey
	LivePeer []*p2pTypes.Peer
	DeadPeer []*p2pTypes.Peer
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

func (server *P2pService) Name() string {
	return "p2p"
}

func (server *P2pService) Api() []app.API {
	return []app.API{}
}

func (server *P2pService) CommandFlags() ([]cli.Command, []cli.Flag) {
	return nil, []cli.Flag{}
}

func (server *P2pService) P2pMessages() map[int]interface{} {
	return map[int]interface{}{
		p2pTypes.MsgTypePing : p2pTypes.Ping{},
		p2pTypes.MsgTypePong : p2pTypes.Pong{},
	}
}

func (server *P2pService) Init(executeContext *app.ExecuteContext) error {
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
	server.Config = &p2pTypes.P2pConfig{}
	err = executeContext.UnmashalConfig(server.Name(), server.Config)
	if err != nil {
		return err
	}

	server.prvKey, err = secp256k1.GeneratePrivateKey(nil)
	if err != nil {
		//TODO shoud never occur
		log.Error("generate private key error ", "Reason", err)
		return err
	}
	server.LivePeer = []*p2pTypes.Peer{}
	server.DeadPeer = []*p2pTypes.Peer{}
	server.inQuene = make(chan *p2pTypes.RouteIn,MaxConnections*2)
	server.outQuene = make(chan *outMessage,MaxConnections*2)
	props := actor.FromProducer(func() actor.Actor {
		return server
	})

	pid, err := actor.SpawnNamed(props, "peer_message")
	if err != nil {
		panic(err)
	}
	server.Router = p2pTypes.NewMsgRouter(server.GetInQuene())
	server.Router.RegisterMsgHandler(p2pTypes.MsgTypePing,pid)
	server.Router.RegisterMsgHandler(p2pTypes.MsgTypePong,pid)
	server.pid = pid
	return nil
}

func (server *P2pService) Start(executeContext *app.ExecuteContext) error {
	server.initBootNodes()
	go server.receiveRoutine()
	go server.sendMessageRoutine()
	go server.recoverDeadPeer()
	return nil
}

func (server *P2pService) Stop(executeContext *app.ExecuteContext) error {
	return nil
}

func (server *P2pService) GetInQuene() chan *p2pTypes.RouteIn{
	return server.inQuene
}

func (server *P2pService) initBootNodes(){
	//init safe
	for _, bootNode := range server.Config.BootNodes {
		if server.isLocalIp(bootNode.IP) {
			continue
		}
		peer := p2pTypes.NewPeer(bootNode.IP,bootNode.Port,bootNode.PubKey,server.handError,server.sendPing)
		if peer.Conn.Connect() {
			peer.Conn.Start()
			server.addPeer(peer)
		} else{
			server.addDeadPeer(peer)
			log.Info("bad boot peer")
		}
	}
}

func (server *P2pService) receiveRoutine(){
	//room for modification addr := &net.TCPAddr{IP: net.ParseIP("x.x.x.x"), Port: receiver.listeningPort()}
	addr := &net.TCPAddr{Port: server.Config.Port}
	if UPnPStart {
		nat.Map("tcp",server.Config.Port, server.Config.Port, "drep nat")
	}

	listener, err := net.ListenTCP("tcp", addr)
	log.Debug("P2p Service started", "addr", listener.Addr())
	if err != nil {
		log.Info("error", err)
		return
	}

	for {
		log.Info("start listen", "port", server.Config.Port)
		conn, err := listener.AcceptTCP()
		log.Info("listen from ", "accept address", conn.RemoteAddr())
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
					msg, msgType, pubkey, err := server.waitForMessage(connTemp)
					if err != nil {
						if err.Error() == "no msg"{
							return
						}else{
							log.Debug("receive message error ","ErrMessage", err.Error())
							return
						}
					}
					peer, err := server.preProcessReq(addr, pubkey)
					if err != nil {
						return
					}
					server.inQuene <- &p2pTypes.RouteIn{
						Type: msgType,
						Peer: peer,
						Detail: msg,
					}
				}(conn)
			}
		}
	}
}

func (server *P2pService) preProcessReq(addr string, pk *secp256k1.PublicKey) (*p2pTypes.Peer, error){
	ipPort := strings.Split(addr,":")
	if server.isLocalIp(ipPort[0]) {
		return nil, errors.New("not allow local ip")
	}
	livePeer := server.SelectPeer(ipPort[0])
	if livePeer == nil {
		deadPeer := server.selectDeadPeer(ipPort[0])
		if deadPeer != nil {
			deadPeer.Conn.ReStart()
			server.addPeer(deadPeer)
			livePeer = deadPeer
		}else {
			livePeer = p2pTypes.NewPeer( ipPort[0],55555,pk,server.handError,server.sendPing)  // //no way to find port
			server.addPeer(livePeer)
		}
	}
	return livePeer, nil
}

func (server *P2pService) waitForMessage(conn *net.TCPConn)(interface{}, int, *secp256k1.PublicKey, error){
	defer conn.Close()
	sizeBytes, err := server.receiveMessageInternal(conn, 4)
	size := (int(sizeBytes[0]) << 24) + (int(sizeBytes[1]) << 16) + (int(sizeBytes[2]) << 8) + int(sizeBytes[3])
	if size == 0 {
		return nil, 0, nil, errors.New("no msg")
	}
	bytes, err := server.receiveMessageInternal(conn, size)
	if err != nil {
		return nil, 0, nil, errors.New("fail to read message ")
	}
	//addr := conn.RemoteAddr().String()
	//log.Debug("receive msg", "Addr", conn.RemoteAddr().String(),"Content", string(bytes))
	return p2pComponent.GetMessage(bytes)
}

func (server *P2pService) receiveMessageInternal(conn net.Conn, size int) ([]byte, error) {
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

func (server *P2pService) handPing(peer *p2pTypes.Peer, ping *p2pTypes.Ping){
	server.SendAsync(peer,&p2pTypes.Pong{})
}

func (server *P2pService) handPong(peer *p2pTypes.Peer, pong *p2pTypes.Pong){
	select {
	case peer.Conn.PongTimeoutCh <- false:
	default:
	}
}

func (server *P2pService) handError(peer *p2pTypes.Peer, err error){
	if err != nil {
		if pErr,ok := err.(*p2pTypes.PeerError);ok {
			log.Error(pErr.Error())
			peer.Conn.Stop()
			server.addDeadPeer(peer)
		}
	}
}


func (server *P2pService) sendPing(peer *p2pTypes.Peer){
	server.SendAsync(peer, &p2pTypes.Ping{})
}

func (server *P2pService) SendMessage(peers []*p2pTypes.Peer, msg interface{}) (sucPeers []*p2pTypes.Peer, failPeers []*p2pTypes.Peer) {
	sucPeers = make([]*p2pTypes.Peer, 0)
	failPeers = make([]*p2pTypes.Peer, 0)
	for _, pk := range peers {
		err := server.Send(pk,msg)
		if err != nil {
			failPeers = append(failPeers, pk)
		} else {
			sucPeers = append(sucPeers, pk)
		}
	}
	return
}

func (server *P2pService) SendAsync(peer *p2pTypes.Peer, msg interface{}){
	server.outQuene <-  &outMessage{Peer:peer, Msg:msg}
}

func (server *P2pService) Send(peer *p2pTypes.Peer, msg interface{}) error{
	done := make(chan error,1)
	server.outQuene <-  &outMessage{Peer:peer, Msg:msg,done:done}
	return <-done
}

func (server *P2pService) Broadcast(msg interface{}){
	for _, peer := range server.LivePeer {
		server.outQuene <-  &outMessage{Peer:peer, Msg:msg}
	}
}

func (server *P2pService) sendMessageRoutine(){
	for {
		select {
		case  outMsg := <-server.outQuene:
			go func() {
				err := server.sendMessage(outMsg)//outMsg.execute()
				if err != nil{
					//dead peer
					server.handError(outMsg.Peer,p2pTypes.NewPeerError(err))
					log.Error("", "MSG",err.Error())
				}
				select {
				case outMsg.done <- err:
				default:
				}
			}()
		case <- server.quit:
			return
		}
	}
}

func (server *P2pService) sendMessage(outMessage *outMessage) error {
	message, err := p2pComponent.GenerateMessage(outMessage.Msg, server.prvKey)
	if err != nil {
		log.Info("error during cipher:", "reason", err)
		return &common.DataError{MyError:common.MyError{Err:err}}
	}
	d, err := time.ParseDuration("3s")
	if err != nil {
		log.Error(err.Error())
		return &common.DefaultError{}
	}
	var conn net.Conn
	for i := 0; i <= 2; i++ {
		conn, err = net.DialTimeout("tcp", outMessage.Peer.GetAddr(), d)
		if err == nil {
			break
		} else {
			log.Info(fmt.Sprintf("%T %v\n", err, err))
			if ope, ok := err.(*net.OpError); ok {
				log.Info(strconv.FormatBool(ope.Timeout()), ope)
			}
			log.Info("Retry after 2s")
			time.Sleep(2 * time.Second)
		}
	}
	if err != nil {
		log.Info(fmt.Sprintf("%T %v\n", err, err))
		if ope, ok := err.(*net.OpError); ok {
			log.Info(strconv.FormatBool(ope.Timeout()), ope)
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
		log.Error(err.Error())
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
		if err := server.sendMessageInternal(conn, sizeBytes); err != nil {
			return &common.TransmissionError{MyError: common.MyError{Err: err}}
		}
		//log.Debug("send message", "IP", conn.RemoteAddr(), "Content", string(bytes))
		if err := server.sendMessageInternal(conn, bytes); err != nil {
			log.Error("Send error ", "Msg", err)
			return &common.TransmissionError{MyError: common.MyError{Err: err}}
		} else {
			return nil
		}
	} else {
		return &common.DataError{MyError:common.MyError{Err:err}}
	}
}

func (server *P2pService) sendMessageInternal(conn net.Conn, bytes []byte) error {
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
func (server *P2pService) recoverDeadPeer(){
	server.tryTimer = time.NewTicker(time.Second * 30)
	for {
		select  {
		case  <-server.tryTimer.C:
			server.peerOpLock.Lock()
			tryPeerCount := 0
			if len(server.DeadPeer) < 40 {    //TODO   MAXPEER * RATE  200*0.2
				tryPeerCount = len(server.DeadPeer)
			} else {
				tryPeerCount = len(server.DeadPeer)/5      //RATE
			}
			tryPeer :=  []*p2pTypes.Peer{}
			for i :=0; i < tryPeerCount; i++ {
				tryPeer = append(tryPeer, server.DeadPeer[i])
			}
			server.peerOpLock.Unlock()
			for _, deadPeer := range tryPeer {
				if deadPeer.Conn.Connect() {
					deadPeer.Conn.ReStart()
					server.addPeer(deadPeer)
					log.Trace("try to connect peer success", "Addr", deadPeer.GetAddr())
				}else{
					deadPeer.Conn.Stop()
					server.addDeadPeer(deadPeer)
					log.Trace("try to connect peer fail", "Addr", deadPeer.GetAddr())
				}
			}
		}
	}
}

func (server *P2pService) Receive(context actor.Context) {
	routeMsg, ok := context.Message().(*p2pTypes.RouteIn)
	if !ok {
		return
	}
	switch msg := routeMsg.Detail.(type) {
	case *p2pTypes.Ping:
		server.handPing(routeMsg.Peer, msg)
	case *p2pTypes.Pong:
		server.handPong(routeMsg.Peer, msg)
	}
}

func (server *P2pService) SelectPeer(ip string)(*p2pTypes.Peer){
	for _,peer := range server.LivePeer {
		if peer.Ip == ip {
			return peer
		}
	}
	return nil
}

func (server *P2pService) selectDeadPeer(ip string)(*p2pTypes.Peer){
	for _,peer := range server.DeadPeer {
		if peer.Ip == ip {
			return peer
		}
	}
	return nil
}

func (server *P2pService) addPeer(peer *p2pTypes.Peer){
	server.peerOpLock.Lock()
	defer server.peerOpLock.Unlock()

	if len(server.LivePeer) > MaxLivePeer {
		return
	}
	server.LivePeer = append(server.LivePeer, peer)

	index := server.indexPeer(server.DeadPeer,peer)
	if index > -1 {
		server.DeadPeer = append(server.DeadPeer[0:index],server.DeadPeer[index+1:len(server.DeadPeer)]...)
	}
}

func (server *P2pService) addDeadPeer(peer *p2pTypes.Peer){
	server.peerOpLock.Lock()
	defer server.peerOpLock.Unlock()

	index :=  server.indexPeer(server.LivePeer,peer)
	if index > -1 {
		server.LivePeer = append(server.LivePeer[0:index],server.LivePeer[index+1:len(server.LivePeer)]...)
	}


	index = server.indexPeer(server.DeadPeer,peer)
	if index == -1 {
		if len(server.DeadPeer) > MaxDeadPeer {
			//remove top peer
			server.DeadPeer = server.DeadPeer[1:len(server.DeadPeer)]
		}
		server.DeadPeer = append(server.DeadPeer,peer)
	}
}

func (server *P2pService) indexPeer(peers []*p2pTypes.Peer,peer *p2pTypes.Peer) int {
	for index, _peer := range  peers {
		if _peer == peer {
			return index
		}
	}
	return -1
}

func (server *P2pService) isLocalIp(ip string) bool{
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

func (server *P2pService) GetIdentifier() *secp256k1.PrivateKey{
	return server.prvKey
}

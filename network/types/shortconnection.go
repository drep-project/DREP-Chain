package types

import (
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"github.com/drep-project/dlog"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

type errorCbFunc func(err error)
type pingCbFunc func()

type PeerError struct {
	err error
}

func NewPeerError(err error) *PeerError {
	return &PeerError{
		err:err,
	}
}
func (pError *PeerError) Error() string{
	return  pError.err.Error()
}

type ShortConnection struct {
	Addr string
	PubKey  *secp256k1.PublicKey
	pingTimer 	  *time.Ticker
	pongTimer     *time.Timer
	PongTimeoutCh chan bool // true - timeout, false - peer sent pong
	handError       errorCbFunc
	handPing        pingCbFunc
	quit chan struct{}
}

func NewShortConnection(addr string, errorHande errorCbFunc, handPing pingCbFunc) *ShortConnection{
	conn := &ShortConnection{
		Addr:addr,
		PongTimeoutCh:make(chan bool,1),
		handError:errorHande,
		handPing: handPing,
	}
	conn.quit = make(chan struct{}, 1)
	return conn
}

func (sConn *ShortConnection) Start() {
	sConn.pingTimer = time.NewTicker(10 * time.Second)
	go sConn.heartbeat()
}

func (sConn *ShortConnection) ReStart() {
	select {			//ensure quit clean
		case <- sConn.quit:
		default:
	}
	sConn.Start()
}

func (sConn *ShortConnection) Stop() {
	if sConn.pingTimer != nil {
		sConn.pingTimer.Stop()
	}
	sConn.stopPongTimer()

	select {
		case sConn.quit <- struct{}{}:
		default:
	}
}

func (sConn *ShortConnection) heartbeat() {
	for{
		select {
		case <- sConn.pingTimer.C:
			sConn.handPing()
			//wait for pong
			sConn.pongTimer = time.AfterFunc(5*time.Second, func() {
				select {
				case sConn.PongTimeoutCh <- true:
				default:
				}
			})
		case timeout := <-sConn.PongTimeoutCh:
			if timeout {
				sConn.Stop()
				sConn.handError(sConn.wrapError(errors.New("Pong timeout")))
			} else {
				sConn.stopPongTimer()
			}
		case <-sConn.quit:
			return
		}
	}
}

func (sConn *ShortConnection) wrapError(err error) error{
	return &PeerError{
		err : err,
	}
}

func (sConn *ShortConnection) stopPongTimer() {
	if sConn.pongTimer != nil {
		_ = sConn.pongTimer.Stop()
		sConn.pongTimer = nil
	}
}

func (sConn *ShortConnection) Connect() bool{
	conn, err := net.DialTimeout("tcp",sConn.Addr, time.Second)
	if err != nil {
		dlog.Error("connection dial fail","Addr",sConn.Addr)
		return false
	}
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	return true
}

func (sConn *ShortConnection) Send(msg []byte) error{
	d, err := time.ParseDuration("3s")
	conn, err := net.DialTimeout("tcp",sConn.Addr, d)
	if err != nil {
		return  err
	}
	defer func() {
		if conn != nil{
			conn.Close()
		}
	}()
	if err != nil {
		dlog.Info(fmt.Sprintf("%T %v\n", err, err))
		if ope, ok := err.(*net.OpError); ok {
			dlog.Info(strconv.FormatBool(ope.Timeout()), ope)
		}
		return err
	}

	now := time.Now()
	d2, err := time.ParseDuration("5s")
	if err != nil {
		dlog.Error(err.Error())
		return err
	} else {
		conn.SetDeadline(now.Add(d2))
	}
	if _, err := conn.Write(msg); err != nil {
		dlog.Info("Send error ", err)
		return err
	}
	return nil
}


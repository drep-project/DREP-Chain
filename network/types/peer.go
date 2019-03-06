package types

import (
	"fmt"
	"github.com/drep-project/drep-chain/crypto/secp256k1"
	"strconv"
	"sync"
)

var (
	DefaultPort = 55555
)

type IP string

func (ip IP) String() string {
	return string(ip)
}

type Port int

func (port Port) String() string {
	return strconv.Itoa(int(port))
}

type Peer struct {
	Ip     string               `json:"ip"`
	Port   int                  `json:"port"`
	PubKey *secp256k1.PublicKey `json:"pubkey"`

	Conn       *ShortConnection `json:"-"`
	addrUpdate sync.Mutex       `json:"-"`
}

func NewPeer(ip string, port int, handError func(*Peer, error), sendPing func(*Peer)) *Peer {
	peer := &Peer{
		Ip:   ip,
		Port: port,
	}
	onError := func(err error) {
		handError(peer, err)
	}
	onPing := func() {
		sendPing(peer)
	}
	peer.Conn = NewShortConnection(peer.GetAddr(), onError, onPing)

	return peer
}

func (peer *Peer) UpdateAddr(ip string, port int) {
	peer.addrUpdate.Lock()
	defer peer.addrUpdate.Unlock()

	peer.Ip = ip
	peer.Port = port
	peer.Conn.Addr = peer.GetAddr()
}

func (peer *Peer) GetAddr() string {
	return fmt.Sprintf("%s:%d", peer.Ip, peer.Port)
}

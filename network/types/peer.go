package types

import (
    "github.com/drep-project/drep-chain/crypto/secp256k1"
    "fmt"
    "strconv"
    "sync"
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
    Ip string
    PubKey  *secp256k1.PublicKey

    Port int

    Conn    *ShortConnection

    addrUpdate sync.Mutex
}

func NewPeer(ip string, port int, pub *secp256k1.PublicKey, handError func(*Peer, error), sendPing func(*Peer)) *Peer {
    peer := &Peer{
        Ip : ip,
        Port: port,
        PubKey: pub,
    }
    onError := func(err error) {
        handError(peer, err)
    }
    onPing := func() {
        sendPing(peer)
    }
    peer.Conn = NewShortConnection(peer.GetAddr(), pub, onError, onPing)

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
    return  fmt.Sprintf("%s:%d",peer.Ip,peer.Port)
}

func DefaultPort() Port{
    return Port(5555)
}
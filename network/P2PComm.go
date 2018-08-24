package network

import (
    "net"
    "sync"
    "BlockChainTest/bean"
)
//const START_BYTE_NORMAL = 0x11
//const START_BYTE_BROADCAST = 0x22

var once sync.Once

type P2PComm struct {
    IPs  []string
    queue  chan *Peer
}
//	RemotePort() int

var sharedInstance *P2PComm

func SharedP2pComm() *P2PComm {
    once.Do(func() {
        sharedInstance = new(P2PComm)
    })
    return sharedInstance
}

func (p *P2PComm) SendMessage(peers []*Peer, msg interface{}) error {
    for _, peer := range peers  {
        // 利用proto buffer序列化
        serializable, err := bean.Serialize(msg)
        if err != nil {
            return err
        }
        p.SendMessageCore(peer, serializable.Body)
    }
    return nil
}

func (p *P2PComm) SendMessageCore(peer *Peer, bytes []byte) error {

    addr, err := net.ResolveTCPAddr("tcp", peer.String())
    if err != nil {
        return err
    }
    conn, err := net.DialTCP("tcp", nil, addr)
    if err != nil {
        return err
    }
    defer conn.Close()
    if _, err := conn.Write(bytes); err != nil {
        return err
    }
    return nil
}

func (p *P2PComm) SpreadMessage() error {
    for _, ip := range p.IPs {
        peer := Peer{ip, ListeningPort,ListeningPort}
        message := Message{peer,p.msg}
        p.queue <- &message
    }
    return nil
}
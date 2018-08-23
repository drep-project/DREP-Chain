package network

import (
    "fmt"
)
//const START_BYTE_NORMAL = 0x11
//const START_BYTE_BROADCAST = 0x22

type P2PComm struct {
    IPs  []string
    port int
    msg  interface{}
    que  chan *Peer
}

var sharedInstance *P2PComm

func (P2PComm) SharedP2pComm() *P2PComm {
    once.Do(func() {
        sharedInstance = new(P2PComm)
    })
    return sharedInstance
}

func (p *P2PComm) SendMessage(peers []*Peer, msg interface{})  {
    for _, peer := range peers  {
        // 利用proto buffer序列化
        _, error := Serialize(msg)
        if error != nil{
            println(error)
            return
        }
        p.SendMessageCore(peer, msg)
    }
}

func (p *P2PComm) SendMessageCore(peer *Peer, msg interface{})  {
    //func NewBroadcast(peer.Net.channel, port int, msg interface{}, que chan *Sender) *Broadcast {
    //return &Broadcast{ips, peer.po, msg, que}
    //}
    peer.initLeader()
    leader := peer.AsLeader
    leader.Listen()
    leader.Work()

    // step 1
    // leader request ticket
    err := leader.RequestTicket()
    if err != nil {
        fmt.Println("error: \n", err)
    }

    if b := leader.ValidateTicket(); b {
        fmt.Println("leader validate ticket: ", b)
        fmt.Println()
    } else {
        return
    }
}

func (p *P2PComm) Spread() error {
    for _, ip := range p.IPs {
        peer := Peer{ip, p.port, p.msg}
        task.BroadcastQueue() <- p
    }
    return nil
}
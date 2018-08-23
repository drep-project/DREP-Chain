package network

import (
    "fmt"
)

type Broadcast struct {
    IPs  []string
    Port int
    Msg  interface{}
    Que  chan *Sender
}

func NewBroadcast(ips []string, port int, msg interface{}, que chan *Sender) *Broadcast {
    return &Broadcast{ips, port, msg, que}
}

func (broadcast *Broadcast) RemoteIPs() []string {
    return broadcast.IPs
}

func (broadcast *Broadcast) RemotePort() int {
    return broadcast.Port
}

func (broadcast *Broadcast) Message() interface{} {
    return broadcast.Msg
}

func (broadcast *Broadcast) BroadcastQueue() chan *Sender {
    return broadcast.Que
}

func (broadcast *Broadcast) Spread() error {
    return SpreadMessages(broadcast)
}

//const START_BYTE_NORMAL = 0x11
//const START_BYTE_BROADCAST = 0x22

type P2PComm struct {
    currentPeer *Peer
}

var sharedInstance *P2PComm

func (P2PComm) SharedP2pComm() *P2PComm {
    once.Do(func() {
        sharedInstance = new(P2PComm)
        sharedInstance.currentPeer = GetPeer()
    })
    return sharedInstance
}

func (p *P2PComm) SendMessage(peers []*Peer, msg interface{})  {
    for _, peer := range peers  {
        // 利用proto buffer序列化
        bytes, error := Serialize(msg)
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
    NewBroadcast(peer)
    GetPeer()
    peer.InitLeader()
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

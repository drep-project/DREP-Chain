package network

import (
    "sync"
    "fmt"
    "time"
)

/*
var (
    once sync.Once
    singleton *Network
)

type Network struct {
    role int
    miningState int
    channel chan *common.Message
}

func GetInstance(channel chan *common.Message) *Network {
    once.Do(func() {
        singleton = new(Network)
        singleton.channel = channel
    })
    return singleton
}

func (n *Network) Start() int {
    go func() {
        for {
            msg := rand.Intn(3)
            time.Sleep(1 * time.Second)
            switch msg {
            case common.MSG_BLOCK:
                n.channel <- &common.Message{common.MSG_BLOCK, common.Block{}}
            case common.MSG_TRANSACTION:
                n.channel <- &common.Message{common.MSG_TRANSACTION, common.Transaction{}}
            }
        }
    }()
    return 0
}
*/

var local = "127.0.0.1"
var LeaderPort = 15287
var MinorPort = 15872
var curve = InitCurve()
var p *Peer
var once sync.Once

type Network struct {
    BroadcastQueue    chan *Sender
    NotificationQueue chan *Notification
}

type Peer struct {
    PrvKey       *PrivateKey
    Net          *Network
    AsLeader     *Leader
    AsMinor      *Minor
}

func GetPeer() *Peer {
    once.Do(func() {
        p = &Peer{}
        p.PrvKey = GetPrvKey()
        p.Net = GetNetwork()
    })
    return p
}

func GetPrvKey() *PrivateKey {
    prvKey, _ := GenerateKey(curve)
    return prvKey
}

func GetNetwork() *Network {
    n := &Network{}
    n.BroadcastQueue = make(chan *Sender, 10)
    n.NotificationQueue = make(chan *Notification, 10)
    return n
}

func GetRoster() ([]string, map[string] *Point) {
    rosterIPs := []string{local}
    roster := make(map[string] *Point)
    roster[local] = GetPrvKey().PubKey
    return rosterIPs, roster
}

func GetPlaintext() interface{} {
    return &CommandOfWord{Msg: []byte("please confirm this block")}
}

func (peer *Peer) InitLeader() error {
    /*
    if peer.AsLeader != nil || peer.AsMinor != nil {
        return errors.New("fail to setup leader, currently involved in another signing protocol")
    }
    */
    leader := &Leader{}
    leader.Word = &CommandOfWord{Msg: []byte("please send your ticket to me")}
    leader.Signal = &SignalOfStart{Mark: 1}
    leader.Plaintext = GetPlaintext()
    leader.RosterIPs, leader.Roster = GetRoster()
    leader.EnterOK = make(map[string] *Ticket)
    leader.CommitOK = make(map[string] *Commitment)
    leader.RespondOK = make(map[string] *Response)
    leader.Net = peer.Net
    peer.AsLeader = leader
    return nil
}

func (peer *Peer) InitMinor() error {
    /*
    if peer.AsLeader != nil || peer.AsMinor != nil {
        return errors.New("fail to setup minor, currently involved in another signing protocol")
    }
    */
    minor := &Minor{}
    minor.PrvKey = peer.PrvKey
    minor.Net = peer.Net
    peer.AsMinor = minor
    return nil
}

func ExecuteMultiSign() {
    peer := GetPeer()
    peer.InitLeader()
    peer.InitMinor()
    leader := peer.AsLeader
    minor := peer.AsMinor
    go leader.Listen()
    leader.Process()
    go minor.Listen()
    minor.Process()

    // step 1
    // leader request ticket
    err := leader.RequestTicket()
    fmt.Println("step 1:")
    fmt.Println("leader request ticket")
    fmt.Println("error: ", err)
    fmt.Println()
    time.Sleep(5 * time.Second)

    // step 2
    // leader request commitment
    err = leader.RequestCommitment()
    fmt.Println("step 2:")
    fmt.Println("leader request commitment")
    fmt.Println("error: ", err)
    fmt.Println()
    time.Sleep(5 * time.Second)

    // validate
    fmt.Println("validate")
    fmt.Println("ticket: ", leader.ValidateTicket())
    fmt.Println("response: ", leader.ValidateResponses())
    fmt.Println()
    time.Sleep(5 * time.Second)
}
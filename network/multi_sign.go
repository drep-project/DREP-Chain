package network

import (
	"errors"
	"net"
	"strings"
	"math/big"
	"math"
    "strconv"
    "time"
    "fmt"
)

type SendUnit interface {
	RemoteIP() string
	RemotePort() int
	Message() interface{}
}

func Address(unit SendUnit) string {
    return unit.RemoteIP() + ":" + strconv.Itoa(unit.RemotePort())
}

func SendMessage(unit SendUnit) error {
	msg, err := Serialize(unit.Message())
	if err != nil {
		return err
	}
	addr, err := net.ResolveTCPAddr("tcp", Address(unit))
	if err != nil {
	    return nil
    }
	conn, err := net.DialTCP("tcp", nil, addr);
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.Write(msg); err != nil {
		return err
	}
	return nil
}

type Sender struct {
	IP   string
	Port int
	Msg  interface{}
}

func NewSender(ip string, port int, msg interface{}) *Sender {
	return &Sender{ip, port, msg}
}

func (sender *Sender) RemoteIP() string {
	return sender.IP
}

func (sender *Sender) RemotePort() int {
	return sender.Port
}

func (sender *Sender) Message() interface{} {
	return sender.Msg
}

func (sender *Sender) Send() error {
	return SendMessage(sender)
}

type MultiSendTask interface {
	RemoteIPs() []string
	RemotePort() int
	Message() interface{}
	BroadcastQueue() chan *Sender
}

func SpreadMessages(task MultiSendTask) error {
	for _, ip := range task.RemoteIPs() {
		sender := NewSender(ip, task.RemotePort(), task.Message())
		task.BroadcastQueue() <- sender
		fmt.Println("broadcast queue: ", len(task.BroadcastQueue()))
	}
	return nil
}

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

type BroadcastReceiver interface {
    ListeningPort() int
    ListeningRole() interface{}
    NotificationQueue() chan *Notification
}

func Listen(receiver BroadcastReceiver) error {
    addr := &net.TCPAddr{Port: receiver.ListeningPort()}
    listener, err := net.ListenTCP("tcp", addr)
    if err != nil {
        return err
    }
    fmt.Println("addr: ", addr)
    for {
        // fmt.Println("listening: ")
        conn, err := listener.AcceptTCP()
        if err != nil {
            continue
        }
        fromAddr := conn.RemoteAddr().String()
        ip := fromAddr[: strings.LastIndex(fromAddr, ":")]
        go func() {
            b := make([]byte, 1024 * 1024)
            buffer := b
            offset := 0
            for {
                n, err := conn.Read(buffer)
                if err != nil {
                    break
                } else {
                    buffer = b[n:]
                    offset += n
                }
            }
            msg, err := Deserialize(b[:offset])
            if err != nil {
                return
            }
            notification := &Notification{receiver.ListeningRole(), msg, ip}
            receiver.NotificationQueue() <- notification
        } ()
    }
}

type Notification struct {
    Role interface{}
    Msg  interface{}
    IP   string
}

func (notification *Notification) Process() error {
    role := notification.Role
    msg := notification.Msg
    ip := notification.IP
    switch role.(type) {
    case *Leader:
        switch msg.(type) {
        case *Ticket:
            role.(*Leader).CollectTicket(msg.(*Ticket), ip)
        case *Commitment:
            role.(*Leader).CollectCommitment(msg.(*Commitment), ip)
        case *Response:
            role.(*Leader).CollectResponse(msg.(*Response), ip)
        default:
            return errors.New("message type not found")
        }
    case *Minor:
        switch msg.(type) {
        case *CommandOfWord:
            role.(*Minor).ReturnTicket(msg.(*CommandOfWord), ip)
        case *SignalOfStart:
            role.(*Minor).ReturnCommitment(msg.(*SignalOfStart), ip)
        case *Challenge:
            role.(*Minor).ReturnResponse(msg.(*Challenge), ip)
        default:
            return errors.New("message type not found")
        }
    default:
        return errors.New("role type not found")
    }
    return nil
}

type Processor interface {
    BroadcastQueue() chan *Sender
    NotificationQueue() chan *Notification
}

func Process(elem interface{}) error {
    switch elem.(type) {
    case *Sender:
        err := elem.(*Sender).Send()
        wait()
        return err
    case *Notification:
        err := elem.(*Notification).Process()
        wait()
        return err
    default:
        return errors.New("wrong queue element type")
    }
}

func Run(processor Processor) {
    go func() {
        for {
            // fmt.Println("processing s: ", len(processor.BroadcastQueue()))
            if len(processor.BroadcastQueue()) > 0 {
                fmt.Println("access broadcast")
                broadcast := <- processor.BroadcastQueue()
                Process(broadcast)
            }
        }
    }()
    go func() {
        for {
            // fmt.Println("processing n: ", len(processor.NotificationQueue()))
            if len(processor.NotificationQueue()) > 0 {
                fmt.Println("access notification")
                notification := <- processor.NotificationQueue()
                Process(notification)
            }
        }
    } ()
}

type Leader struct {
	Word            *CommandOfWord
	Signal          *SignalOfStart
	Chn             *Challenge
	Plaintext       interface{}
	RosterIPs       []string
	Roster          map[string] *Point
	EnterOK         map[string] *Ticket
	CommitOK        map[string] *Commitment
	RespondOK       map[string] *Response
	Net             *Network
}

func (leader *Leader) RequestTicket() error {
    return NewBroadcast(leader.RosterIPs, MinorPort, leader.Word, leader.Net.BroadcastQueue).Spread()
}

func (leader *Leader) CollectTicket(ticket *Ticket, ip string) {
    if leader.EnterOK == nil {
        leader.EnterOK = make(map[string] *Ticket)
    }
    leader.EnterOK[ip] = ticket
}

func (leader *Leader) ValidateTicket() bool {
    if len(leader.EnterOK) != len(leader.Roster) {
        return false
    }
    for ip, p := range leader.Roster {
        if q, b := leader.EnterOK[ip]; !b {
            return false
        } else if !PointEqual(p, q.PubKey) {
            return false
        } else if !Verify(curve, q.SigOfWord, q.PubKey, leader.Word.Msg) {
            return false
        }
    }
    return true
}

func (leader *Leader) RequestCommitment() error {
    return NewBroadcast(leader.RosterIPs, MinorPort, leader.Signal, leader.Net.BroadcastQueue).Spread()
}

func (leader *Leader) CollectCommitment(commitment *Commitment, ip string) {
    if leader.CommitOK == nil {
        leader.CommitOK = make(map[string] *Commitment)
    }
    leader.CommitOK[ip] = commitment
}

func (leader *Leader) ProposeChallenge() error {
    msg, err := Serialize(leader.Plaintext)
    if err != nil {
        return err
    }
    groupPubKey := &Point{}
    groupQ := &Point{}
    ips := make([]string, 0)
    for ip, commitment := range leader.CommitOK {
        ips = append(ips, ip)
        groupPubKey = curve.Add(groupPubKey, commitment.PubKey)
        groupQ = curve.Add(groupQ, commitment.Q)
    }
    r := ConcatEncode(groupQ, groupPubKey, msg)
    challenge := &Challenge{GroupPubKey: groupPubKey, GroupQ: groupQ, Msg: msg, R: r}
    leader.Chn = challenge
    return NewBroadcast(ips, MinorPort, challenge, leader.Net.BroadcastQueue).Spread()
}

func (leader *Leader) CollectResponse(response *Response, ip string) {
    if leader.RespondOK == nil {
        leader.RespondOK = make(map[string] *Response)
    }
    leader.RespondOK[ip] = response
}

func (leader *Leader) ValidateResponses() bool {
    if len(leader.RespondOK) < len(leader.CommitOK) {
        return false
    }
    if float64(len(leader.RespondOK)) < math.Ceil(float64(len(leader.Roster) * 2.0 / 3.0) + 1) {
        return false
    }
    groupPubKey := &Point{}
    groupS := new(big.Int)
    for ip, response := range leader.RespondOK {
        if q, b := leader.CommitOK[ip]; !b {
            return false
        } else if !PointEqual(response.PubKey, q.PubKey) {
            return false
        }
        groupPubKey = curve.Add(groupPubKey, response.PubKey)
        groupS.Add(groupS, new(big.Int).SetBytes(response.S))
    }
    groupS.Mod(groupS, curve.N)
    sig := &Signature{R: leader.Chn.R, S: groupS.Bytes()}
    msg, err := Serialize(leader.Plaintext)
    if err != nil {
        return false
    }
    return Verify(curve, sig, groupPubKey, msg)
}

func (leader *Leader) ListeningPort() int {
    return LeaderPort
}

func (leader *Leader) ListeningRole() interface{} {
    return leader
}

func (leader *Leader) BroadcastQueue() chan *Sender {
    return leader.Net.BroadcastQueue
}

func (leader *Leader) NotificationQueue() chan *Notification {
    return leader.Net.NotificationQueue
}

func (leader *Leader) Listen() error {
    return Listen(leader)
}

func (leader *Leader) Process() error {
    Run(leader)
    return nil
}

type Minor struct {
	PrvKey *PrivateKey
	K      []byte
	Net    *Network
}

func (minor *Minor) ReturnTicket(word *CommandOfWord, ip string) error {
    ips := append(make([]string, 0), ip)
    sigOfWord, err := Sign(curve, minor.PrvKey, []byte(word.Msg))
    if err != nil {
        return err
    }
    ticket := &Ticket{}
    ticket.PubKey = minor.PrvKey.PubKey
    ticket.SigOfWord = sigOfWord
    return NewBroadcast(ips, LeaderPort, ticket, minor.Net.BroadcastQueue).Spread()
}

func (minor *Minor) ReturnCommitment(signal *SignalOfStart, ip string) error {
    k, err := RandomSample(curve)
    if err != nil {
        return err
    }
    minor.K = k
    ips := append(make([]string, 0), ip)
    commitment := &Commitment{PubKey: minor.PrvKey.PubKey, Q: curve.ScalarBaseMultiply(k)}
    return NewBroadcast(ips, LeaderPort, commitment, minor.Net.BroadcastQueue).Spread()
}

func (minor *Minor) ReturnResponse(challenge *Challenge, ip string) error {
    r := ConcatEncode(challenge.GroupQ, challenge.GroupPubKey, challenge.Msg)
    r0 := new(big.Int).SetBytes(challenge.R)
    r1 := new(big.Int).SetBytes(r)
    if r0.Cmp(r1) != 0 {
        return errors.New("wrong hash value")
    }
    kInt := new(big.Int).SetBytes(minor.K)
    prvInt := new(big.Int).SetBytes(minor.PrvKey.Prv)
    s := new(big.Int).Mul(r1, prvInt)
    s.Sub(kInt, s)
    s.Mod(s, curve.N)
    response := &Response{PubKey: minor.PrvKey.PubKey, S: s.Bytes()}
    ips := append(make([]string, 0), ip)
    return NewBroadcast(ips, LeaderPort, response, minor.Net.BroadcastQueue).Spread()
}

func (minor *Minor) ListeningPort() int {
    return MinorPort
}

func (minor *Minor) ListeningRole() interface{} {
    return minor
}

func (minor *Minor) BroadcastQueue() chan *Sender {
    return minor.Net.BroadcastQueue
}

func (minor *Minor) NotificationQueue() chan *Notification {
    return minor.Net.NotificationQueue
}

func (minor *Minor) Listen() error {
    return Listen(minor)
}

func (minor *Minor) Process() error {
    Run(minor)
    return nil
}

func wait() {
    time.Sleep(1 * time.Second)
}
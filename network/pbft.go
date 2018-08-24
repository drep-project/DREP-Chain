package network
//import (
//	"errors"
//	"net"
//	"strings"
//	"math/big"
//	"math"
//   "strconv"
//   "fmt"
//)
//
//type SendUnit interface {
//	RemoteIP() string
//	RemotePort() int
//	Message() interface{}
//}
//
//func RemoteAddress(unit SendUnit) string {
//   return unit.RemoteIP() + ":" + strconv.Itoa(unit.RemotePort())
//}
//
//func LocalAddress() string {
//   return local
//}
//
//func SendMessage(unit SendUnit) error {
//	msg, err := Serialize(unit.Message())
//
//	if err != nil {
//		return err
//	}
//	addr, err := net.ResolveTCPAddr("tcp", RemoteAddress(unit))
//	if err != nil {
//	    return nil
//   }
//	conn, err := net.DialTCP("tcp", nil, addr)
//	if err != nil {
//		return err
//	}
//	defer conn.Close()
//	if _, err := conn.Write(msg); err != nil {
//		return err
//	}
//	return nil
//}
//
//type Sender struct {
//	IP   string
//	Port int
//	Msg  interface{}
//}
//
//func NewSender(ip string, port int, msg interface{}) *Sender {
//	return &Sender{ip, port, msg}
//}
//
//func (sender *Sender) RemoteIP() string {
//	return sender.IP
//}
//
//func (sender *Sender) RemotePort() int {
//	return sender.Port
//}
//
//func (sender *Sender) Message() interface{} {
//	return sender.Msg
//}
//
//func (sender *Sender) Send() error {
//	return SendMessage(sender)
//}
//
//type MultiSendTask interface {
//	RemoteIPs() []string
//	RemotePort() int
//	Message() interface{}
//	BroadcastQueue() chan *Sender
//}
//
//func SpreadMessages(task MultiSendTask) error {
//	for _, ip := range task.RemoteIPs() {
//		sender := NewSender(ip, task.RemotePort(), task.Message())
//		task.BroadcastQueue() <- sender
//	}
//	return nil
//}
//
//type Sender struct {
//	IPs  []string
//	Port int
//	Msg  interface{}
//	Que  chan *Sender
//}
//
//func NewBroadcast(ips []string, port int, msg interface{}, que chan *Sender) *Sender {
//	return &Sender{ips, port, msg, que}
//}
//
//func (broadcast *Sender) RemoteIPs() []string {
//	return broadcast.IPs
//}
//
//func (broadcast *Sender) RemotePort() int {
//	return broadcast.Port
//}
//
//func (broadcast *Sender) Message() interface{} {
//	return broadcast.Msg
//}
//
//func (broadcast *Sender) BroadcastQueue() chan *Sender {
//   return broadcast.Que
//}
//
//func (broadcast *Sender) Spread() error {
//	return SpreadMessages(broadcast)
//}
//
//type BroadcastReceiver interface {
//   ListeningPort() int
//   ListeningRole() interface{}
//   NotificationQueue() chan *Notification
//}
//
//func Listen(receiver BroadcastReceiver) {
//   go func() {
//
//       // addr := &net.TCPAddr{IP: net.ParseIP("172.20.10.6"), Port: receiver.ListeningPort()}
//       addr := &net.TCPAddr{Port: receiver.ListeningPort()}
//       listener, err := net.ListenTCP("tcp", addr)
//       if err != nil {
//           return
//       }
//       for {
//           conn, err := listener.AcceptTCP()
//           if err != nil {
//               continue
//           }
//           fromAddr := conn.RemoteAddr().String()
//           ip := fromAddr[: strings.LastIndex(fromAddr, ":")]
//           go func() {
//               b := make([]byte, 1024)
//               buffer := b
//               offset := 0
//               for {
//                   n, err := conn.Read(buffer)
//                   if err != nil {
//                       break
//                   } else {
//                       buffer = b[n:]
//                       offset += n
//                   }
//               }
//               msg, err := Deserialize(b[:offset])
//               if err != nil {
//                   return
//               }
//               notification := &Notification{receiver.ListeningRole(), msg, ip}
//               receiver.NotificationQueue() <- notification
//           } ()
//       }
//   }()
//}
//
//type Notification struct {
//   Role interface{}
//   Msg  interface{}
//   IP   string
//}
//
//func (notification *Notification) Process() error {
//   role := notification.Role
//   msg := notification.Msg
//   ip := notification.IP
//   switch role.(type) {
//   case *Leader:
//       switch msg.(type) {
//       case *Ticket:
//           fmt.Println("leader process ticket")
//           fmt.Println()
//           role.(*Leader).CollectTicket(msg.(*Ticket), ip)
//       case *Commitment:
//           fmt.Println("leader process commitment")
//           fmt.Println()
//           role.(*Leader).CollectCommitment(msg.(*Commitment), ip)
//       case *Response:
//           fmt.Println("leader process response")
//           fmt.Println()
//           role.(*Leader).CollectResponse(msg.(*Response), ip)
//       default:
//           return errors.New("message type not found")
//       }
//   case *Minor:
//       switch msg.(type) {
//       case *CommandOfWord:
//           fmt.Println("minor process command of word")
//           fmt.Println()
//           role.(*Minor).ReturnTicket(msg.(*CommandOfWord), ip)
//       case *SignalOfStart:
//           fmt.Println("minor process signal of start")
//           fmt.Println()
//           role.(*Minor).ReturnCommitment(msg.(*SignalOfStart), ip)
//       case *Challenge:
//           fmt.Println("minor process challenge")
//           fmt.Println()
//           role.(*Minor).ReturnResponse(msg.(*Challenge), ip)
//       default:
//           return errors.New("message type not found")
//       }
//   default:
//       return errors.New("role type not found")
//   }
//   return nil
//}
//
//type Processor interface {
//   BroadcastQueue() chan *Sender
//   NotificationQueue() chan *Notification
//}
//
//func Process(elem interface{}) error {
//   switch elem.(type) {
//   case *Sender:
//       err := elem.(*Sender).Send()
//       return err
//   case *Notification:
//       err := elem.(*Notification).Process()
//       return err
//   default:
//       return errors.New("wrong queue element type")
//   }
//}
//
//func Work(processor Processor) {
//   go func() {
//       for {
//           if broadcast, ok := <- processor.BroadcastQueue(); ok {
//               Process(broadcast)
//           }
//       }
//   }()
//   go func() {
//       for {
//           if notification, ok := <- processor.NotificationQueue(); ok {
//               Process(notification)
//           }
//       }
//   } ()
//}
//
//type Leader struct {
//	Command            *CommandOfWord
//	Announcement          *SignalOfStart
//	Challenge             *Challenge
//	Object       interface{}
//	MinorAddresses       []string
//	MinorPubKeys          map[string] *Point
//	EnterOK         map[string] *Ticket
//	CommitOK        map[string] *Commitment
//	RespondOK       map[string] *Response
//	Net             *Network
//}
//
//func (leader *Leader) RequestTicket() error {
//   return NewBroadcast(leader.MinorAddresses, MinorPort, leader.Command, leader.Net.BroadcastQueue).Spread()
//}
//
//func (leader *Leader) CollectTicket(ticket *Ticket, ip string) {
//   if leader.EnterOK == nil {
//       leader.EnterOK = make(map[string] *Ticket)
//   }
//   leader.EnterOK[ip] = ticket
//}
//
//func (leader *Leader) ValidateTicket() bool {
//   if len(leader.EnterOK) != len(leader.MinorPubKeys) {
//       return false
//   }
//   for ip, p := range leader.MinorPubKeys {
//       if q, b := leader.EnterOK[ip]; !b {
//           return false
//       } else if !PointEqual(p, q.PubKey) {
//           return false
//       } else if !Verify(curve, q.SigOfWord, q.PubKey, leader.Command.Msg) {
//           return false
//       }
//   }
//   return true
//}
//
//func (leader *Leader) RequestCommitment() error {
//   return NewBroadcast(leader.MinorAddresses, MinorPort, leader.Announcement, leader.Net.BroadcastQueue).Spread()
//}
//
//func (leader *Leader) CollectCommitment(commitment *Commitment, ip string) {
//   if leader.CommitOK == nil {
//       leader.CommitOK = make(map[string] *Commitment)
//   }
//   leader.CommitOK[ip] = commitment
//}
//
//func (leader *Leader) ProposeChallenge() error {
//   msg, err := Serialize(leader.Object)
//   if err != nil {
//       return err
//   }
//   groupPubKey := &Point{}
//   groupQ := &Point{}
//   ips := make([]string, 0)
//   for ip, commitment := range leader.CommitOK {
//       ips = append(ips, ip)
//       groupPubKey = curve.Add(groupPubKey, commitment.PubKey)
//       groupQ = curve.Add(groupQ, commitment.Q)
//   }
//   r := ConcatHash(groupQ, groupPubKey, msg)
//   challenge := &Challenge{GroupPubKey: groupPubKey, GroupQ: groupQ, Msg: msg, R: r}
//   leader.Challenge = challenge
//   return NewBroadcast(ips, MinorPort, challenge, leader.Net.BroadcastQueue).Spread()
//}
//
//func (leader *Leader) CollectResponse(response *Response, ip string) {
//   if leader.RespondOK == nil {
//       leader.RespondOK = make(map[string] *Response)
//   }
//   leader.RespondOK[ip] = response
//}
//
//func (leader *Leader) ValidateResponses() bool {
//   if len(leader.RespondOK) < len(leader.CommitOK) {
//       return false
//   }
//   if float64(len(leader.RespondOK)) < math.Ceil(float64(len(leader.MinorPubKeys) * 2.0 / 3.0) + 1) {
//       return false
//   }
//   groupPubKey := &Point{}
//   groupS := new(big.Int)
//   for ip, response := range leader.RespondOK {
//       if q, b := leader.CommitOK[ip]; !b {
//           return false
//       } else if !PointEqual(response.PubKey, q.PubKey) {
//           return false
//       }
//       groupPubKey = curve.Add(groupPubKey, response.PubKey)
//       groupS.Add(groupS, new(big.Int).SetBytes(response.S))
//   }
//   groupS.Mod(groupS, curve.N)
//   sig := &Signature{R: leader.Challenge.R, S: groupS.Cipher()}
//   msg, err := Serialize(leader.Object)
//   if err != nil {
//       return false
//   }
//   return Verify(curve, sig, groupPubKey, msg)
//}
//
//func (leader *Leader) ListeningPort() int {
//   return LeaderPort
//}
//
//func (leader *Leader) ListeningRole() interface{} {
//   return leader
//}
//
//func (leader *Leader) BroadcastQueue() chan *Sender {
//   return leader.Net.BroadcastQueue
//}
//
//func (leader *Leader) NotificationQueue() chan *Notification {
//   return leader.Net.NotificationQueue
//}
//
//func (leader *Leader) Listen() error {
//   Listen(leader)
//   return nil
//}
//
//func (leader *Leader) Work() error {
//   Work(leader)
//   return nil
//}
//
//type Minor struct {
//	PrvKey *PrivateKey
//	K      []byte
//	Net    *Network
//}
//
//func (minor *Minor) ReturnTicket(word *CommandOfWord, ip string) error {
//   ips := append(make([]string, 0), ip)
//   sigOfWord, err := Sign(curve, minor.PrvKey, []byte(word.Msg))
//   if err != nil {
//       return err
//   }
//   ticket := &Ticket{}
//   ticket.PubKey = minor.PrvKey.PubKey
//   ticket.SigOfWord = sigOfWord
//   return NewBroadcast(ips, LeaderPort, ticket, minor.Net.BroadcastQueue).Spread()
//}
//
//func (minor *Minor) ReturnCommitment(signal *SignalOfStart, ip string) error {
//   k, err := GetRandomKQ(curve)
//   if err != nil {
//       return err
//   }
//   minor.K = k
//   ips := append(make([]string, 0), ip)
//   commitment := &Commitment{PubKey: minor.PrvKey.PubKey, Q: curve.ScalarBaseMultiply(k)}
//   return NewBroadcast(ips, LeaderPort, commitment, minor.Net.BroadcastQueue).Spread()
//}
//
//func (minor *Minor) ReturnResponse(challenge *Challenge, ip string) error {
//   r := ConcatHash(challenge.GroupQ, challenge.GroupPubKey, challenge.Msg)
//   r0 := new(big.Int).SetBytes(challenge.R)
//   r1 := new(big.Int).SetBytes(r)
//   if r0.Cmp(r1) != 0 {
//       return errors.New("wrong hash value")
//   }
//   kInt := new(big.Int).SetBytes(minor.K)
//   prvInt := new(big.Int).SetBytes(minor.PrvKey.Prv)
//   s := new(big.Int).Mul(r1, prvInt)
//   s.Sub(kInt, s)
//   s.Mod(s, curve.N)
//   response := &Response{PubKey: minor.PrvKey.PubKey, S: s.Cipher()}
//   ips := append(make([]string, 0), ip)
//   return NewBroadcast(ips, LeaderPort, response, minor.Net.BroadcastQueue).Spread()
//}
//
//func (minor *Minor) ListeningPort() int {
//   return MinorPort
//}
//
//func (minor *Minor) ListeningRole() interface{} {
//   return minor
//}
//
//func (minor *Minor) BroadcastQueue() chan *Sender {
//   return minor.Net.BroadcastQueue
//}
//
//func (minor *Minor) NotificationQueue() chan *Notification {
//   return minor.Net.NotificationQueue
//}
//
//func (minor *Minor) Listen() error {
//   Listen(minor)
//   return nil
//}
//
//func (minor *Minor) Work() error {
//   Work(minor)
//   return nil
//}
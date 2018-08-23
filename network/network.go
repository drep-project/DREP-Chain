package network

import (
    "time"
    "sync"
    "fmt"
    "errors"
    "net"
    "BlockChainTest/common"
   "strconv"
   "proj/crypto"
   "strings"
)

//var (
//    once sync.Once
//    singleton *Network
//)
//
//type Network struct {
//    role int
//    miningState int
//    channel chan *common.Message
//}
//
//func GetInstance(channel chan *common.Message) *Network {
//    once.Do(func() {
//        singleton = new(Network)
//        singleton.channel = channel
//    })
//    return singleton
//}
//
//func (n *Network) Start() int {
//    go func() {
//        for {
//            msg := rand.Intn(3)
//            time.Sleep(1 * time.Second)
//            switch msg {
//            case common.MSG_BLOCK:
//                n.channel <- &common.Message{common.MSG_BLOCK, common.Block{rand.Intn(1000), "Block"}}
//            case common.MSG_TRANSACTION:
//                n.channel <- &common.Message{common.MSG_TRANSACTION, common.Transaction{rand.Intn(1000), "Transaction"}}
//            }
//        }
//    }()
//    return 0
//}

//var local = "127.0.0.1"
var local = "124.78.94.72"
var LeaderIP = "124.78.94.72"
var MinorIP = "192.168.3.73"
var LeaderPort = 15287
var MinorPort = 56595
var NonMinorPort = 17263
var curve = InitCurve()
var CheckFrequency = 50 * time.Nanosecond
var MaximumWaitTime = 500 * time.Nanosecond
var p *Peer
var key *PrivateKey
var once0, once1 sync.Once

type Network struct {
   BroadcastQueue    chan *Sender
   NotificationQueue chan *Notification
}

type Task struct {
}

var once0, once1 sync.Once
var SenderQueue chan *Sender
var ReceiverQueue chan *Receiver

type IP string

func (ip IP) String() string {
   return string(ip)
}

type Port int

func (port Port) String() string {
   return strconv.Itoa(int(port))
}

func URL(ip IP, port Port) string {
   return ip.String() + ":" + port.String()
}

type Address string

func (addr Address) String() string {
   return string(addr)
}

func (addr Address) LocalKey() string {
   return ADDRESS_SUFFIX + addr.String()
}

type Sender struct {
   RemoteIP   IP
   RemotePort Port
   Msg        interface{}
}

func GetSenderQueue() chan *Sender {
   once0.Do(func() {
      SenderQueue = make(chan *Sender,  10)
   })
   return SenderQueue
}

func (sender *Sender) Send() error {
   msg, err := common.Serialize(sender.Msg)
   if err != nil {
   		return err
   }
   addr, err := net.ResolveTCPAddr("tcp", URL(sender.RemoteIP, sender.RemotePort))
   if err != nil {
     return nil
   }
   conn, err := net.DialTCP("tcp", nil, addr)
   if err != nil {
     return err
   }
   defer conn.Close()
   if _, err := conn.Write(msg); err != nil {
      return err
   }
   return nil
}

func SendMessage(peers []Peer, msg interface{}) {
   queue := GetSenderQueue()
   for _, addr := range addresses {
      sender := &Sender{Addr, msg}
      queue <- sender
   }
}

type Receiver struct {
   Addr Address
   Msg interface{}
}

func Listen() {
  go func() {
     //room for modification addr := &net.TCPAddr{IP: net.ParseIP("x.x.x.x"), Port: receiver.ListeningPort()}
     addr := &net.TCPAddr{Port: ListeningPort}
     listener, err := net.ListenTCP("tcp", addr)
     if err != nil {
        return
     }
     for {
        conn, err := listener.AcceptTCP()
        if err != nil {
           continue
        }
        fromAddr := conn.RemoteAddr().String()
        ip := fromAddr[: strings.LastIndex(fromAddr, ":")]
        go func() {
           b := make([]byte, BufferSize)
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
  }()
}

//type NonMinor struct {
//    DB  *database.Database
//    Net *Network
//}
//
//func (nom *NonMinor) ListeningPort() int {
//    return NonMinorPort
//}
//
//func (nom *NonMinor) ListeningRole() interface{} {
//    return nom
//}
//
//func (nom *NonMinor) BroadcastQueue() chan *Sender {
//    return nom.Net.BroadcastQueue
//}
//
//func (nom *NonMinor) NotificationQueue() chan *Notification {
//    return nom.Net.NotificationQueue
//}
//
//func (nom *NonMinor) Listen() error {
//    Listen(nom)
//    return nil
//}
//
//func (nom *NonMinor) Work() error {
//    Work(nom)
//    return nil
//}

type Peer struct {
   RemoteIP IP
   RemotePort Port
}

func GetPeer() *Peer {
   once0.Do(func() {
       p = &Peer{}
       p.PrvKey = GetPrvKey()
       p.Net = GetNetwork()
   })
   return p
}

func GetPrvKey() *PrivateKey {
   once1.Do(func() {
       var prvKey *PrivateKey = nil
       err := errors.New("fail to generate key pair")
       for err != nil {
           prvKey, err = GenerateKey(curve)
           key = prvKey
           time.Sleep(100 * time.Millisecond)
       }
   })
   return key
}

func GetNetwork() *Network {
   n := &Network{}
   n.BroadcastQueue = make(chan *Sender, 10)
   n.NotificationQueue = make(chan *Notification, 10)
   return n
}

func GetRoster() ([]string, map[string] *Point) {
   rosterIPs := []string{MinorIP}
   roster := make(map[string] *Point)
   roster[MinorIP] = GetPrvKey().PubKey
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
   //minor := peer.AsMinor
   leader.Listen()
   leader.Work()
   //minor.Listen()
   //minor.Work()

   // step 1
   // leader request ticket
   err := leader.RequestTicket()
   fmt.Println("step 1:")
   fmt.Println("leader request ticket")
   fmt.Println("error: ", err)
   fmt.Println()

   var wait int64 = 0
   for wait < MaximumWaitTime.Nanoseconds() {
      wait += 1
      time.Sleep(time.Nanosecond)
   }

   if b := leader.ValidateTicket(); b {
      fmt.Println("leader validate ticket: ", b)
      fmt.Println()
   } else {
      return
   }

   // step 2
   // leader request commitment
   err = leader.RequestCommitment()
   fmt.Println("step 2:")
   fmt.Println("leader request commitment")
   fmt.Println("error: ", err)
   fmt.Println()

   wait = 0
   for wait < MaximumWaitTime.Nanoseconds() {
      wait += 1
      time.Sleep(time.Nanosecond)
   }

   // step 3
   // leader propose challenge
   err = leader.ProposeChallenge()
   fmt.Println("step 3:")
   fmt.Println("leader propose challenge")
   fmt.Println("error: ", err)
   fmt.Println()

   wait = 0
   for wait < MaximumWaitTime.Nanoseconds() {
      wait += 1
      time.Sleep(time.Nanosecond)
   }

   if b := leader.ValidateResponses(); b {
      fmt.Println("leader validate response: ", b)
      fmt.Println()
   } else {
      return
   }
}
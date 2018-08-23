package network

import (
   "sync"
   "strconv"
   "BlockChainTest/common"
   "net"
   "strings"
)

var onceSender, onceReceiver sync.Once
var SenderQueue chan *Message
var ReceiverQueue chan *Message

type IP string

func (ip IP) String() string {
   return string(ip)
}

type Port int

func (port Port) String() string {
   return strconv.Itoa(int(port))
}

type Address string

func (addr Address) String() string {
   return string(addr)
}

func (addr Address) LocalKey() string {
   return ADDRESS_SUFFIX + addr.String()
}

type Peer struct {
   RemoteIP IP
   RemotePort Port
}

func (peer *Peer) String() string {
   return peer.RemoteIP.String() + ":" + peer.RemotePort.String()
}

type Message struct {
   RemotePeer *Peer
   Msg interface{}
}

func GetSenderQueue() chan *Message {
   onceSender.Do(func() {
      SenderQueue = make(chan *Message,  10)
   })
   return SenderQueue
}

func GetReceiverQueue() chan *Message {
   onceReceiver.Do(func() {
      ReceiverQueue = make(chan *Message, 10)
   })
   return ReceiverQueue
}

func (m *Message) Send() error {
   msg, err := common.Serialize(m.Msg)
   if err != nil {
   		return err
   }
   addr, err := net.ResolveTCPAddr("tcp", m.RemotePeer.String())
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

func SendMessage(peers []*Peer, msg interface{}) {
   queue := GetSenderQueue()
   for _, peer := range peers {
      message := &Message{peer, msg}
      queue <- message
   }
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
        ip := fromAddr[:strings.LastIndex(fromAddr, ":")]
        peer := &Peer{RemoteIP: IP(ip), RemotePort: ListeningPort}
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
              msg, err := common.Deserialize(b[:offset])
              if err != nil {
                 return
              }
              message := &Message{peer, msg}
              queue := GetReceiverQueue()
              queue <- message
           }
        }()
     }
  }()
}
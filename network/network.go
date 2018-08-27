package network

import (
   "sync"
   "strconv"
   "net"
   "strings"
   "BlockChainTest/bean"
   "github.com/golang/protobuf/proto"
   "BlockChainTest/crypto"
   "errors"
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
   return AddressSuffix + addr.String()
}

type Peer struct {
   RemoteIP IP
   RemotePort Port
   RemotePubKey *bean.Point
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

func (m *Message) Cipher() ([]byte, error) {
   serializable, err := bean.Serialize(m.Msg)
   if err != nil {
      return nil, err
   }
   sig, err := crypto.Sign(serializable.Body)
   if err != nil {
      return nil, err
   }
   serializable.Sig = sig
   pubKey, err := crypto.GetPubKey()
   if err != nil {
      return nil, err
   }
   serializable.PubKey = pubKey
   plaintext, err := proto.Marshal(serializable)
   if err != nil {
      return nil, err
   }
   cipher, err := crypto.Encrypt(m.RemotePeer.RemotePubKey, plaintext)
   if err != nil {
      return nil, err
   }
   return cipher, nil
}

func (m *Message) Send() error {
   cipher, err := m.Cipher()
   if err != nil {
      return err
   }
   addr, err := net.ResolveTCPAddr("tcp", m.RemotePeer.String())
   if err != nil {
     return err
   }
   conn, err := net.DialTCP("tcp", nil, addr)
   if err != nil {
     return err
   }
   defer conn.Close()
   if _, err := conn.Write(cipher); err != nil {
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

func DecryptIntoMessage(cipher []byte) (*Message, error) {
   plaintext, err := crypto.Decrypt(cipher)
   if err != nil {
      return nil, err
   }
   serializable, msg, err := bean.Deserialize(plaintext)
   if err != nil {
      return nil, err
   }
   if !crypto.Verify(serializable.Sig, serializable.PubKey, serializable.Body) {
      return nil, errors.New("decrypt fail")
   }
   peer := &Peer{RemotePubKey: serializable.PubKey}
   message := &Message{RemotePeer: peer, Msg: msg}
   return message, nil
}

func Listen(process func(int, interface{})) {
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
        b := make([]byte, BufferSize)
        cipher := b
        offset := 0
        for {
           n, err := conn.Read(cipher)
           if err != nil {
              break
           } else {
              cipher = b[n:]
              offset += n
           }
           message, err := DecryptIntoMessage(cipher)
           if err != nil {
              return
           }
           fromAddr := conn.RemoteAddr().String()
           ip := fromAddr[:strings.LastIndex(fromAddr, ":")]
           message.RemotePeer.RemoteIP = IP(ip)
           //queue := GetReceiverQueue()
           //queue <- message
           //p := processor.GetInstance()
           t, msg := bean.IdentifyMessage(message)
           if msg != nil {
              process(t, msg)
           }
        }
     }
  }()
}

func Work() {
   go func() {
      sender := GetSenderQueue()
      for {
         if message, ok := <-sender; ok {
            message.Send()
         }
      }
   }()
}
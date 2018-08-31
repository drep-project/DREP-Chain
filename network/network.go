package network

import (
   "sync"
   "net"
   "fmt"
   "strings"
)

var onceSender sync.Once
var SenderQueue chan *Message

func GetSenderQueue() chan *Message {
   onceSender.Do(func() {
      SenderQueue = make(chan *Message,  10)
   })
   return SenderQueue
}

func SendMessage(peers []*Peer, msg interface{}) {
   queue := GetSenderQueue()
   for _, peer := range peers {
      message := &Message{peer, msg}
      queue <- message
   }
}

func startListen(process func(int, interface{})) {
  go func() {
     //room for modification addr := &net.TCPAddr{IP: net.ParseIP("x.x.x.x"), Port: receiver.listeningPort()}
     addr := &net.TCPAddr{Port: listeningPort}
     listener, err := net.ListenTCP("tcp", addr)
     if err != nil {
        fmt.Println("error", err)
        return
     }
     for {
        fmt.Println("start listen")
        conn, err := listener.AcceptTCP()
        fmt.Println("listen from ", conn.RemoteAddr())
        if err != nil {
           continue
        }
        cipher := make([]byte, bufferSize)
        b := cipher
        offset := 0
        for {
           n, err := conn.Read(b)
           if err != nil {
              break
           } else {
              offset += n
              b = cipher[offset:]
           }
        }
        fmt.Println("Receive ", cipher[:offset])
        fmt.Println("Receive byte ", offset)
        message, err := DecryptIntoMessage(cipher[:offset])
        fmt.Println("Receive after decrypt", message)
        if err != nil {
           return
        }
        fromAddr := conn.RemoteAddr().String()
        ip := fromAddr[:strings.LastIndex(fromAddr, ":")]
        message.Peer.IP = IP(ip)
        //queue := GetReceiverQueue()
        //queue <- message
        //p := processor.GetInstance()
        t, msg := identifyMessage(message)
        if msg != nil {
           process(t, msg)
        }
        fmt.Println("end listen")
     }
  }()
}

func startSend() {
   go func() {
      sender := GetSenderQueue()
      for {
         if message, ok := <-sender; ok {
            message.Send()
         }
      }
   }()
}

func Start(process func(int, interface{})) {
   startListen(process)
   startSend()
}
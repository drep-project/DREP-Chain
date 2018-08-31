package network

import (
   "sync"
   "net"
   "fmt"
   "strings"
)

var onceSender sync.Once
var SenderQueue chan *Task

func GetSenderQueue() chan *Task {
   onceSender.Do(func() {
      SenderQueue = make(chan *Task,  10)
   })
   return SenderQueue
}

func SendMessage(peers []*Peer, msg interface{}) {
   queue := GetSenderQueue()
   for _, peer := range peers {
      task := &Task{peer, msg}
      queue <- task
   }
}

//func SendMessage(peers []*Peer, msg interface{}) error {
//    for _, peer := range peers  {
//        //use proto buffer serialize
//        serializable, err := bean.Serialize(msg)
//        if err != nil {
//            return err
//        }
//        SendMessageCore(peer, serializable.Body)
//    }
//    return nil
//}
//
//func SendMessageCore(peer *Peer, bytes []byte) error {
//
//    addr, err := net.ResolveTCPAddr("tcp", peer.ToString())
//    if err != nil {
//        return err
//    }
//    conn, err := net.DialTCP("tcp", nil, addr)
//    if err != nil {
//        return err
//    }
//    defer conn.Close()
//    if _, err := conn.Write(bytes); err != nil {
//        return err
//    }
//    return nil
//}

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
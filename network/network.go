package network

import (
   "sync"
   "net"
   "strings"
    "BlockChainTest/mycrypto"
    "BlockChainTest/bean"
    "BlockChainTest/log"
)

const (
    listeningPort = 55555
    bufferSize    = 1024 * 1024
)

var onceSender sync.Once
var SenderQueue chan *Task

func getSenderQueue() chan *Task {
   onceSender.Do(func() {
      SenderQueue = make(chan *Task,  10)
   })
   return SenderQueue
}

func SendMessage(peers []*Peer, msg interface{}) {
   queue := getSenderQueue()
   for _, peer := range peers {
      task := &Task{peer, msg}
      queue <- task
   }
}

func Start(process func(int, interface{})) {
    startListen(process)
    startSend()
}

func startListen(process func(int, interface{})) {
    go func() {
        //room for modification addr := &net.TCPAddr{IP: net.ParseIP("x.x.x.x"), Port: receiver.listeningPort()}
        addr := &net.TCPAddr{Port: listeningPort}
        listener, err := net.ListenTCP("tcp", addr)
        if err != nil {
            log.Println("error", err)
            return
        }
        for {
            log.Println("start listen")
            conn, err := listener.AcceptTCP()
            log.Println("listen from ", conn.RemoteAddr())
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
            log.Println("Receive ", cipher[:offset])
            log.Println("Receive byte ", offset)
            task, err := decryptIntoTask(cipher[:offset])
            log.Println("Receive after decrypt", task)
            if err != nil {
                return
            }
            fromAddr := conn.RemoteAddr().String()
            ip := fromAddr[:strings.LastIndex(fromAddr, ":")]
            task.Peer.IP = IP(ip)
            //queue := GetReceiverQueue()
            //queue <- message
            //p := processor.GetInstance()
            t, msg := identifyMessage(task)
            if msg != nil {
                process(t, msg)
            }
            log.Println("end listen")
        }
    }()
}

func startSend() {
   go func() {
      sender := getSenderQueue()
      for {
         if task, ok := <-sender; ok {
            log.Println(task.Peer.IP)
            task.execute()
         }
      }
   }()
}

func decryptIntoTask(cipher []byte) (*Task, error) {
    plaintext, err := mycrypto.Decrypt(cipher)
    if err != nil {
        return nil, err
    }
    serializable, msg, err := bean.Deserialize(plaintext)
    if err != nil {
        return nil, err
    }
    //if !mycrypto.Verify(serializable.Sig, serializable.PubKey, serializable.Body) {
    //   return nil, errors.New("decrypt fail")
    //}
    peer := &Peer{PubKey: serializable.PubKey}
    task := &Task{Peer: peer, Msg: msg}
    return task, nil
}

func identifyMessage(task *Task) (int, interface{}) {
    msg := task.Msg
    switch msg.(type) {
    case *bean.Setup:
        return bean.MsgTypeSetUp, msg.(*bean.Setup)
    case *bean.Commitment:
        return bean.MsgTypeCommitment, msg.(*bean.Commitment)
    case *bean.Challenge:
        return bean.MsgTypeChallenge, msg.(*bean.Challenge)
    case *bean.Response:
        return bean.MsgTypeResponse, msg.(*bean.Response)
    case *bean.Block:
        return bean.MsgTypeBlock, msg.(*bean.Block)
    case *bean.PeerInfo:
        return bean.MsgTypeNewPeer, msg.(*bean.PeerInfo)
    case *bean.PeerInfoList:
        return bean.MsgTypePeerList,msg.(*bean.PeerInfoList)
    default:
        return -1, nil
    }
}
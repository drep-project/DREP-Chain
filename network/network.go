package network

import (
    "sync"
    "net"
    "strings"
    "BlockChainTest/mycrypto"
    "BlockChainTest/bean"
    //"BlockChainTest/log"
    "BlockChainTest/util"
    "BlockChainTest/network/nat"
    "errors"
    "BlockChainTest/store"
    "fmt"
)

const (
    bufferSize    = 1024 * 1024
    UPnPStart  = false
)

var (
    lock sync.Mutex
)

func SendMessage(peers []*bean.Peer, msg interface{}) (sucPeers []*bean.Peer, failPeers []*bean.Peer) {
   lock.Lock()
   defer lock.Unlock()
   sucPeers = make([]*bean.Peer, 0)
   failPeers = make([]*bean.Peer, 0)
   for _, peer := range peers {
       task := &Task{PrvKey:store.GetPrvKey(), Peer:peer, Msg:msg}
       if err := task.execute(); err != nil {
           switch err.(type) {
           case *util.TimeoutError, *util.ConnectionError:
               failPeers = append(failPeers, peer)
           }
       } else {
           sucPeers = append(sucPeers, peer)
       }
   }
   return
}

func Start(process func(*bean.Peer, int, interface{}), port bean.Port) {
    startListen(process, port)
}

func startListen(process func(*bean.Peer, int, interface{}), port bean.Port) {
    go func() {
        //room for modification addr := &net.TCPAddr{IP: net.ParseIP("x.x.x.x"), Port: receiver.listeningPort()}
        addr := &net.TCPAddr{Port: int(port)}
        if UPnPStart {
            nat.Map("tcp", int(port), int(port), "drep nat")
        }
        listener, err := net.ListenTCP("tcp", addr)
        if err != nil {
            fmt.Println("error", err)
            return
        }
        for {
            fmt.Println("start listen", port)
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
            task, err := decryptIntoTask(cipher[:offset]) // TODO what the fuck is this???
            fmt.Println("Receive after decrypt", task)
            if err != nil {
                return
            }
            fromAddr := conn.RemoteAddr().String()
            ip := fromAddr[:strings.LastIndex(fromAddr, ":")]
            task.Peer.IP = bean.IP(ip)
            //queue := GetReceiverQueue()
            //queue <- message
            //p := processor.GetInstance()
            t, msg := identifyMessage(task)
            if msg != nil {
                process(task.Peer, t, msg)
            }
            fmt.Println("end listen")
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
    if !mycrypto.Verify(serializable.Sig, serializable.PubKey, serializable.Body) {
      return nil, errors.New("wrong signature")
    }
    peer := &bean.Peer{PubKey: serializable.PubKey}
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
    case *bean.Transaction:
        return bean.MsgTypeTransaction, msg.(*bean.Transaction)
    case *bean.BlockReq:
        return bean.MsgTypeBlockReq, msg.(*bean.BlockReq)
    case *bean.BlockResp:
        return bean.MsgTypeBlockResp, msg.(*bean.BlockResp)
    case *bean.Ping:
        return bean.MsgTypePing, msg.(*bean.Ping)
    case *bean.Pong:
        return bean.MsgTypePong, msg.(*bean.Pong)
    case *bean.OfflinePeers:
        return bean.MsgTypeOfflinePeers, msg.(*bean.OfflinePeers)
    case *bean.FirstPeerInfoList:
        return bean.MsgTypeFirstPeerInfoList, msg.(*bean.FirstPeerInfoList)
    default:
        return -1, nil
    }
}

func GetIps() []string {
    r := make([]string, 0)
    if addrs, err := net.InterfaceAddrs(); err == nil {
        for _, a := range addrs {
            if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
                if ipnet.IP.To4() != nil {
                    r = append(r, ipnet.IP.String())
                }
            }
        }
    }
    return r
}
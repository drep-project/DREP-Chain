package network

//import (
//    "strconv"
//    "net"
//    "errors"
//    "time"
//    "math/big"
//    "fmt"
//    "BlockChainTest/bean"
//)
//
//var serverIP = "127.0.0.1"
//var portPort = 14767
//var serverPubKey *bean.Point
//var requestNeighboursWord = "please send me my neighbours' ips"
//
//var ReadFrequency = 100 * time.Nanosecond
//var ReadTimeout = 10000 * time.Nanosecond
//var Room = 10
//
//func Read(conn *net.TCPConn, size int) ([]byte, error) {
//    buffer := make([]byte, size)
//    offset := 0
//    var dur time.Duration = 0
//    for offset < size {
//        n, err := conn.Read(buffer[offset:])
//        offset += n
//        if err != nil {
//            break
//        }
//        time.Sleep(ReadFrequency)
//        dur += ReadFrequency
//        if dur > ReadTimeout {
//            return nil, errors.New("read time out")
//        }
//    }
//    if offset < size {
//        return nil, errors.New("read insufficient bytes")
//    }
//    return buffer, nil
//}
//
//func Write(conn *net.TCPConn, object interface{}) error {
//    defer conn.CloseWrite()
//    b, err := bean.Marshal(object)
//    if err != nil {
//        return err
//    }
//    size := len(b)
//    buffer := make([]byte, 4)
//    copy(buffer, new(big.Int).SetInt64(int64(size)).Bytes())
//    if _, err := conn.Write(buffer); err != nil {
//        return err
//    }
//    if _, err := conn.Write(b); err != nil {
//        return err
//    }
//    return nil
//}
//
//func OrderedRead(conn *net.TCPConn) []interface{} {
//   defer conn.CloseRead()
//   list := make([]interface{}, Room)
//   num := 0
//   var err error = nil
//   var objectSize []byte
//   var objectBytes []byte
//   for err == nil {
//       if len(list) == num {
//           list = append(list, make([]interface{}, Room))
//           objectSize, err = Read(conn, 4)
//           if err != nil {
//               break
//           }
//           size := int(new(big.Int).SetBytes(objectSize).Int64())
//           if size == 0 {
//               break
//           }
//           objectBytes, err = Read(conn, size)
//           if err != nil {
//               break
//           }
//           object, err := Deserialize(objectBytes)
//           if err != nil {
//               break
//           }
//           list[num] = object
//           num ++
//       }
//   }
//   return list
//}
//
//func OrderedWrite(conn *net.TCPConn, objects []interface{}) error {
//    defer conn.CloseWrite()
//    for _, object := range objects {
//        err := Write(conn, object)
//        if err != nil {
//            fmt.Println("error: ", err)
//        }
//    }
//    return nil
//}
//
//type Link interface {
//    LinkingIP() string
//    LinkingPort() int
//}
//
//func GetAddress(link Link) string {
//    return link.LinkingIP() + ":" + strconv.Itoa(link.LinkingPort())
//}
//
//func GetConn(link Link) (*net.TCPConn, error) {
//    addr, err := net.ResolveTCPAddr("tcp", GetAddress(link))
//    if err != nil {
//        return nil, err
//    }
//    conn, err := net.DialTCP("tcp", nil, addr);
//    if err != nil {
//        return nil, err
//    }
//    return conn, nil
//}
//
//type NonMinor struct {
//    IP string
//    Port int
//    PrvKey *PrivateKey
//    TargetPubKey *Point
//    //DB *database.Database
//}
//
//func (nom *NonMinor) Connect() (*net.TCPConn, error) {
//    return GetConn(nom)
//}
//
//func (nom *NonMinor) Fetch() {
//    conn, err := nom.Connect()
//    if err != nil {
//        return
//    }
//    defer conn.Close()
//    cipher := Encrypt(curve, nom.TargetPubKey, []byte("please send me my neighbours"));
//    if cipher == nil {
//        return
//    }
//    sig, err := Sign(curve, nom.PrvKey, cipher)
//    if err != nil {
//        return
//    }
//    fetchRequest := &FetchRequest{FetchWhat: cipher, Sig: sig}
//    b, err := Serialize(fetchRequest)
//    if err != nil {
//        return
//    }
//    n, err := conn.Write(b)
//    fmt.Println("n: ", n)
//    if err != nil {
//        return
//    }
//    list := OrderedRead(conn)
//    // process list
//    fmt.Println("list: ", list)
//    return
//}
//
//func (nom *NonMinor) AssignNeighbours() error {
//    return nil
//}
//
//func (nom *NonMinor) SendBlocks(height int) error {
//    conn, err := nom.Connect()
//    if err != nil {
//        return err
//    }
//    ch := height
//    var block *Block
//    blockKey := "block_" + strconv.Itoa(height)
//    block = nom.GetBlock(blockKey)
//    for block != nil {
//        Write(conn, block)
//        ch += 1
//        blockKey = "block_" + strconv.Itoa(ch)
//        block = nom.GetBlock(blockKey)
//    }
//    return nil
//}
//
//func (nom *NonMinor) GetBlock(blockKey string) *Block {
//    return &Block{}
//}
//
//func (nom *NonMinor) LinkingIP() string {
//    return nom.IP
//}
//
//func (nom *NonMinor) LinkingPort() int {
//    return nom.Port
//}
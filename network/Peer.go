package network

import (
    "BlockChainTest/bean"
    "strconv"
)

//var key *common.PrivateKey

type IP string

func (ip IP) String() string {
    return string(ip)
}

type Port int

func (port Port) String() string {
    return strconv.Itoa(int(port))
}

type Peer struct {
    IP      IP
    Port    Port
    PubKey  *bean.Point
    Address bean.Address
}

func (peer *Peer) String() string {
    return peer.IP.String() + ":" + peer.Port.String()
}

//func GetPrvKey() *PrivateKey {
//   once1.Do(func() {
//       var prvKey *PrivateKey = nil
//       err := errors.New("fail to generate key pair")
//       for err != nil {
//           prvKey, err = GenerateKey(curve)
//           key = prvKey
//           time.Sleep(100 * time.Millisecond)
//       }
//   })
//   return key
//}

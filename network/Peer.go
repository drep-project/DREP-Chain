package network

import (
    "strconv"
    "errors"
    "time"
    "sync"
)

var local = "127.0.0.1"
var curve = InitCurve()
var key *PrivateKey
var once0, once1 sync.Once

type Peer struct {
    IP       string
    msg      interface{}
}


func (peer *Peer) String() string{
    return peer.IP + ":" + strconv.Itoa(peer.Port)
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

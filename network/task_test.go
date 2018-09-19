package network

import (
    "testing"
    "BlockChainTest/mycrypto"
    "fmt"
)

func InitPeer() *Peer {
    curve := mycrypto.GetCurve()
    k := []byte{0x22, 0x11}
    pub := curve.ScalarBaseMultiply(k)
    ip := IP("192.168.3.73")
    port := Port(55555)
    peer := &Peer{IP: ip, Port: port, PubKey: pub}
    return peer
}

func  TestTask_Cipher(t *testing.T) {
    peer := InitPeer()
    msg := "this is a message"
    task := Task{peer, msg}
    bytes, err := task.cipher()
    if err != nil {
        if bytes != nil {
            fmt.Println(bytes)
        } else  {
            fmt.Println("msg is nill")
        }
        t.Log("test func cipher is passed, and the bytes is", bytes)
    } else  {
        t.Error("there is an error during testing the func cipher")
    }
}

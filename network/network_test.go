package network

import (
    "testing"
    "BlockChainTest/mycrypto"
)

func InitPeers() []*Peer {
    curve := mycrypto.GetCurve()
    k0 := []byte{0x22, 0x11}
    k1 := []byte{0x14, 0x44}
    k2 := []byte{0x11, 0x55}
    pub0 := curve.ScalarBaseMultiply(k0)
    pub1 := curve.ScalarBaseMultiply(k1)
    pub2 := curve.ScalarBaseMultiply(k2)

    ip0 := IP("192.168.3.13")
    ip1 := IP("192.168.3.43")
    ip2 := IP("192.168.3.73")
    port0 := Port(55555)
    port1 := Port(55555)
    port2 := Port(55555)
    peer0 := &Peer{IP: ip0, Port: port0, PubKey: pub0}
    peer1 := &Peer{IP: ip1, Port: port1, PubKey: pub1}
    peer2 := &Peer{IP: ip2, Port: port2, PubKey: pub2}
    return []*Peer{peer0, peer1, peer2}
}

func TestSendMessage(t *testing.T) {
    peers := InitPeers()
    SendMessage(peers,"this is a msg")
}

func TestDecryptIntoMessage(t *testing.T) {

}
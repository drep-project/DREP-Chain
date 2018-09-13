package network

import (
    "BlockChainTest/bean"
    "strconv"
    "BlockChainTest/mycrypto"
)

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
    PubKey  *mycrypto.Point
    Address bean.Address
}

func (peer *Peer) ToString() string {
    return peer.IP.String() + ":" + peer.Port.String()
}
package network

import (
    "strconv"
    "net"
    "BlockChainTest/database"
)

var serverIP = "127.0.0.1"
var portPort = 14767
var serverPubKey *Point
var requestNeighboursWord = "please send me my neighbours' ips"

type Link interface {
    LinkingIP() string
    LinkingPort() int
}

func GetAddress(link Link) string {
    return link.LinkingIP() + ":" + strconv.Itoa(link.LinkingPort())
}

func GetConn(link Link) (*net.TCPConn, error) {
    addr, err := net.ResolveTCPAddr("tcp", GetAddress(link))
    if err != nil {
        return nil, err
    }
    conn, err := net.DialTCP("tcp", nil, addr);
    if err != nil {
        return nil, err
    }
    return conn, nil
}

type NonMinor struct {
    IP string
    Port int
    PrvKey PrivateKey
    Local *database.Database
}

func (nom *NonMinor) Connect() (*net.TCPConn, error) {
    return GetConn(nom)
}

func (nom *NonMinor) RequestNeighbours() error {
    conn, err := nom.Connect()
    if err != nil {
        return err
    }
    defer conn.Close()
    // cip, err := Sign(curve, )
    // sig, err := Sign(curve, serverPubKey)
    return nil
}

func (nom *NonMinor) LinkingIP() string {
    return nom.IP
}

func (nom *NonMinor) LinkingPort() int {
    return nom.Port
}
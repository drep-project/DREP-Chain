package network

import "net"

type Peer struct {
    net.IP
    //port
}

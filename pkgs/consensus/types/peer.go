package types

import (
	"github.com/drep-project/drep-chain/network/p2p"
)

var (
	DefaultPort = 55555
)

type IPeerInfo interface {
	GetMsgRW() p2p.MsgReadWriter
	String() string
	IP() string
	Equal(ipeer IPeerInfo) bool
}

//业务层peer
type PeerInfo struct {
	peer        *p2p.Peer
	rw          p2p.MsgReadWriter
}

func NewPeerInfo(peer *p2p.Peer, rw p2p.MsgReadWriter) *PeerInfo {
	return &PeerInfo{
		peer: peer,
		rw:   rw,
	}
}

//获取读写句柄
func (pi *PeerInfo) GetMsgRW() p2p.MsgReadWriter {
	return pi.rw
}

func (pi *PeerInfo) IP() string {
	return pi.peer.IP()
}


func (pi *PeerInfo) String() string {
	return pi.peer.IP()
}

func (pi *PeerInfo) Equal(pi2  IPeerInfo) bool {
	return pi2.IP() == pi.IP()
}
package types

import (
	"github.com/deckarep/golang-set"
	"github.com/drep-project/drep-chain/network/p2p"
	"github.com/drep-project/drep-chain/network/p2p/enode"
)

var (
	DefaultPort = 55555
)

//业务层peer
type PeerInfo struct {
	knownBlocks mapset.Set // Set of block hashes known to be known by this peer
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

func (pi *PeerInfo) GetID() *enode.ID {
	id := pi.peer.ID()
	return &id
}

func (pi *PeerInfo) IP() string {
	return pi.peer.IP()
}
